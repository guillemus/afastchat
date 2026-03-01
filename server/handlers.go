package server

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"

	"datastar-chat/db"
	openaiPkg "datastar-chat/openai"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) appRoutes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("POST /chat/send", s.handleSendMessage)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := s.renderer.Render(w, "index", nil); err != nil {
		slog.Error("render index", "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

type sendSignals struct {
	Message  string `json:"message"`
	ThreadID string `json:"threadId"`
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var sig sendSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Warn("readSignals failed", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	msg := strings.TrimSpace(sig.Message)
	if msg == "" {
		slog.Warn("empty message received")
		http.Error(w, "empty message", http.StatusBadRequest)
		return
	}

	slog.Info("message received", "len", len(msg), "threadId", sig.ThreadID)

	ctx := r.Context()

	// Resolve or create thread
	var threadID uuid.UUID
	if sig.ThreadID != "" {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			slog.Warn("invalid thread id", "threadId", sig.ThreadID, "err", err)
			http.Error(w, "invalid thread id", http.StatusBadRequest)
			return
		}
		threadID = parsed
	} else {
		thread, err := s.db.CreateThread(ctx, "New Chat")
		if err != nil {
			slog.Error("create thread", "err", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		threadID = thread.ID
		slog.Info("created thread", "threadId", threadID)
	}

	// Save user message
	userMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ThreadID: threadID,
		Role:     "user",
		Content:  msg,
	})
	if err != nil {
		slog.Error("create user message", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Create placeholder assistant message
	assistantMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ThreadID: threadID,
		Role:     "assistant",
		Content:  "",
	})
	if err != nil {
		slog.Error("create assistant message", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	slog.Info("db messages saved", "userMsgId", userMsg.ID, "assistantMsgId", assistantMsg.ID)

	// Open SSE connection (MUST be after ReadSignals)
	sse := datastar.NewSSE(w, r, datastar.WithCompression())

	// Patch signals: set threadId, clear message, set loading
	sse.MarshalAndPatchSignals(map[string]any{
		"threadId": threadID.String(),
		"message":  "",
		"loading":  true,
	})

	userDivID := fmt.Sprintf("msg-%s", userMsg.ID)
	assistantDivID := fmt.Sprintf("msg-%s", assistantMsg.ID)

	// Append user + assistant message bubbles before scroll anchor
	sse.PatchElements(
		fmt.Sprintf(
			`<div id="%s" class="message user-message">%s</div>
			 <div id="%s" class="message assistant-message"></div>`,
			userDivID, html.EscapeString(msg), assistantDivID,
		),
		datastar.WithSelectorID("scroll-anchor"),
		datastar.WithModeBefore(),
	)

	// Build messages for OpenAI from DB history
	dbMessages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		slog.Error("get messages", "threadId", threadID, "err", err)
		return
	}
	slog.Info("loaded history", "threadId", threadID, "msgCount", len(dbMessages))

	oaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(dbMessages))
	for _, m := range dbMessages {
		if m.Content == "" {
			continue
		}
		switch m.Role {
		case "user":
			oaiMessages = append(oaiMessages, openai.UserMessage(m.Content))
		case "assistant":
			oaiMessages = append(oaiMessages, openai.AssistantMessage(m.Content))
		case "system":
			oaiMessages = append(oaiMessages, openai.SystemMessage(m.Content))
		}
	}

	// Stream from OpenAI
	slog.Info("starting openai stream", "threadId", threadID, "model", "gpt-5.2", "oaiMsgCount", len(oaiMessages))
	oaiStream := openaiPkg.Client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    "gpt-5.2",
		Messages: oaiMessages,
	})

	var accumulated strings.Builder
	for oaiStream.Next() {
		chunk := oaiStream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		token := chunk.Choices[0].Delta.Content
		if token == "" {
			continue
		}
		accumulated.WriteString(token)

		sse.PatchElements(
			html.EscapeString(token),
			datastar.WithSelectorID(assistantDivID),
			datastar.WithModeAppend(),
		)
	}

	if err := oaiStream.Err(); err != nil {
		slog.Error("openai stream", "threadId", threadID, "err", err)
	}
	oaiStream.Close()

	// Save final content
	finalContent := accumulated.String()
	slog.Info("stream complete", "threadId", threadID, "contentLen", len(finalContent))
	if err := s.db.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
		ID:      assistantMsg.ID,
		Content: finalContent,
	}); err != nil {
		slog.Error("update assistant message", "err", err)
	}

	// Signal loading done
	sse.MarshalAndPatchSignals(map[string]any{
		"loading": false,
	})
}
