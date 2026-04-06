package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"datastar-chat/agents"
	"datastar-chat/db"
	"datastar-chat/views"

	"github.com/google/uuid"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	threads, err := s.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	s.renderer.RenderIndex(w, views.IndexData{Threads: threads})
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request, threadID uuid.UUID) {
	ctx := r.Context()

	thread, err := s.db.GetThread(ctx, threadID)
	if err != nil {
		handleBadReq(w, "get thread", err)
		return
	}

	messages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		handleBadReq(w, "get messages", err)
		return
	}

	for i := range messages {
		if messages[i].Role == "assistant" {
			messages[i].Content = renderMarkdown(messages[i].Content)
		}
	}

	threads, err := s.db.ListThreads(ctx)
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	s.renderer.RenderThread(w, views.ThreadData{
		Threads:   threads,
		Thread:    thread,
		Messages:  messages,
		Streaming: thread.IsStreaming,
	})
}

func (s *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request, threadID uuid.UUID) {
	if err := s.db.ArchiveThread(r.Context(), threadID); err != nil {
		handleBadReq(w, "archive thread", err)
		return
	}

	sse := datastar.NewSSE(w, r)

	sse.RemoveElementByID(fmt.Sprintf("thread-%s", threadID))
	sse.ExecuteScript(fmt.Sprintf(
		`if (window.location.pathname === '/threads/%s') window.location.replace('/')`,
		threadID,
	))
}

type SearchThreadsSig struct {
	ThreadSearch string `json:"threadSearch"`
	ThreadID     string `json:"threadId"`
}

func (s *Server) handleSearchThreads(w http.ResponseWriter, r *http.Request, sig SearchThreadsSig) {
	threads, err := s.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	threads, err = s.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	query := sig.ThreadSearch
	filtered := make([]db.Thread, 0, len(threads))
	for _, t := range threads {
		if query == "" || strings.Contains(strings.ToLower(t.Title), query) {
			filtered = append(filtered, t)
		}
	}

	var activeThreadID uuid.UUID
	if sig.ThreadID != "" {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			slog.Warn("invalid threadId signal", "threadId", sig.ThreadID, "err", err)
		} else {
			activeThreadID = parsed
		}
	}

	var threadsSb strings.Builder
	views.RenderThreads(&threadsSb, filtered, activeThreadID)

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(
		threadsSb.String(),
		datastar.WithSelectorID("thread-items"),
		datastar.WithModeInner(),
	)
}

type SendMessageSignals struct {
	Message  string     `json:"message"`
	ThreadID *uuid.UUID `json:"threadId"`
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request, sig SendMessageSignals) {
	ctx := r.Context()

	userMessage := sig.Message

	var threadID uuid.UUID
	isNewThread := sig.ThreadID == nil
	if isNewThread {
		id, err := s.db.CreateThreadWithMessages(ctx,
			"New Chat", "You are a helpful assistant", userMessage)
		if err != nil {
			handleBadReq(w, "create thread", err)
			return
		}
		slog.Info("created thread", "threadId", threadID)
		threadID = *id
	} else {
		_, err := s.db.GetThread(ctx, *sig.ThreadID)
		if errors.Is(err, sql.ErrNoRows) {
			handleBadReq(w, "thread not found", err)
			return
		} else if err != nil {
			handleBadReq(w, "get thread", err)
			return
		}

		threadID = *sig.ThreadID
	}

	if isNewThread {
		slog.Info("creating new thread")

		go func() {
			defer recoverAndLog()
			s.updateThreadTitle(context.Background(), threadID, userMessage)
		}()

		sse := datastar.NewSSE(w, r)

		slog.Info("redirecting")
		sse.Redirectf("/threads/%s", threadID)
	}

	go func() {
		defer recoverAndLog()
		s.writeAgentStreamToThread(threadID, userMessage)
	}()
}

func (s *Server) writeAgentStreamToThread(threadID uuid.UUID, userMessage string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID)

	// Set streaming state
	if err := s.db.SetThreadStreaming(ctx, threadID, true); err != nil {
		logger.Error("set streaming state", "err", err)
		return
	}
	defer func() {
		if err := s.db.SetThreadStreaming(ctx, threadID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
		}
	}()

	messages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	messages = append(messages, db.Message{
		ThreadID:  threadID,
		Role:      "user",
		Content:   userMessage,
		CreatedAt: time.Now(),
	})

	logger.Info("starting agent stream")
	agent := agents.NewAgent()

	// create empty assistant message so that we have an id to go and update.
	assistantMsgID, err := s.db.CreateAssistantMessage(ctx, threadID, "")
	if err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var assistantMsg strings.Builder

	for token, err := range agent.Stream(ctx, messages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		assistantMsg.WriteString(token)

		publishThreadMessage(s.bus, threadID, assistantMsg.String())

		select {
		case <-ticker.C:
			err := s.db.UpdateMessageContent(ctx, assistantMsgID, assistantMsg.String())
			if err != nil {
				logger.Error("update assistant message", "err", err)
			}
		default:
		}
	}

	// Final update with complete message
	err = s.db.UpdateMessageContent(ctx, assistantMsgID, assistantMsg.String())
	if err != nil {
		logger.Error("final update assistant message", "err", err)
	}
}

func (s *Server) updateThreadTitle(ctx context.Context, threadID uuid.UUID, userMessage string) {
	agent := agents.NewAgent()
	title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

	err := s.db.UpdateThreadTitle(ctx, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}
}

func (s *Server) handleThreadEvents(w http.ResponseWriter, r *http.Request, threadID uuid.UUID) {
	ctx := r.Context()

	messagesC := make(chan string)
	handler := func(message string) {
		messagesC <- message
	}

	unsub := subscribeToThread(s.bus, threadID, handler)
	defer unsub()

	// create datastar sse with brotli compression
	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli()))

msgloop:
	for {
		select {
		case msg := <-messagesC:
			html := renderMarkdown(msg)
			var sb strings.Builder
			views.RenderAssistantMessage(&sb, threadID, html)

			sse.PatchElements(sb.String())
		case <-ctx.Done():
			break msgloop
		}
	}
}
