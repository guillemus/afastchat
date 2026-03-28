package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"datastar-chat/agent"
	"datastar-chat/db"
	"datastar-chat/views"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) renderThreadItem(thread db.Thread, activeThreadID uuid.UUID) string {
	return s.renderer.RenderThreadItem(thread, activeThreadID)
}

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

	_, streaming := s.streams.Get(threadID)
	slog.Info("handleThread", "threadId", threadID, "streaming", streaming)

	s.renderer.RenderThread(w, views.ThreadData{
		Threads:   threads,
		Thread:    thread,
		Messages:  messages,
		Streaming: streaming,
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
		`if (window.location.pathname === '/threads/%s') { window.location.replace('/') }`,
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

	var threadItemsHTML strings.Builder
	for _, t := range filtered {
		threadItemHTML := s.renderThreadItem(t, activeThreadID)
		threadItemsHTML.WriteString(threadItemHTML)
	}

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(
		threadItemsHTML.String(),
		datastar.WithSelectorID("thread-items"),
		datastar.WithModeInner(),
	)
}

type SendMessageSignals struct {
	Message  string `json:"message"`
	ThreadID string `json:"threadId"`
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request, sig SendMessageSignals) {
	ctx := r.Context()

	slog.Info("message received", "threadId", sig.ThreadID)

	// Resolve or create thread
	isNewThread := sig.ThreadID == ""
	var threadID uuid.UUID
	if !isNewThread {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			handleBadReq(w, "invalid thread id", err)
			return
		}
		threadID = parsed
	} else {
		thread, err := s.db.CreateThread(ctx, db.CreateThreadParams{
			ID:    uuid.New(),
			Title: "New Chat",
		})
		if err != nil {
			handleBadReq(w, "create thread", err)
			return
		}
		threadID = thread.ID
		slog.Info("created thread", "threadId", threadID)
	}

	// Save user message
	userMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "user",
		Content:  sig.Message,
	})
	if err != nil {
		slog.Error("create user message", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Create placeholder assistant message
	assistantMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "assistant",
		Content:  "",
	})
	if err != nil {
		handleBadReq(w, "create assistant message", err)
		return
	}

	slog.Info("db messages saved", "userMsgId", userMsg.ID, "assistantMsgId", assistantMsg.ID)

	// Register stream immediately so refresh can resume even if user reloads right away.
	stream := newActiveStream(threadID, assistantMsg.ID)
	s.streams.Set(threadID, stream)

	// Open SSE connection (MUST be after ReadSignals)
	sse := datastar.NewSSE(w, r, datastar.WithCompression())

	// Patch signals: set threadId, clear message, set loading
	sse.MarshalAndPatchSignals(map[string]any{
		"threadId": threadID.String(),
		"message":  "",
		"loading":  true,
	})

	// If new thread, prepend to sidebar
	if isNewThread {
		thread := db.Thread{ID: threadID, Title: "New Chat"}

		threadItemHTML := s.renderThreadItem(thread, threadID)
		sse.PatchElements(
			threadItemHTML,
			datastar.WithSelectorID("thread-items"),
			datastar.WithModePrepend(),
		)

		sse.ExecuteScript(fmt.Sprintf(
			`history.pushState({}, '', '/threads/%s')`,
			threadID,
		))
	}

	assistantDivID := fmt.Sprintf("msg-%s", assistantMsg.ID)

	// Append user + assistant message bubbles before scroll anchor
	userMessageHTML := views.RenderUserMessage(userMsg.ID, sig.Message, true)
	assistantMessageHTML := views.RenderAssistantMessage(assistantMsg.ID, "", true)

	sse.PatchElements(
		userMessageHTML+assistantMessageHTML,
		datastar.WithSelectorID("scroll-anchor"),
		datastar.WithModeBefore(),
	)

	// Build messages for OpenAI from DB history
	dbMessages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		slog.Error("get messages", "threadId", threadID, "err", err)
		stream.Finish()
		s.streams.DeleteIfSame(threadID, stream)
		sse.MarshalAndPatchSignals(map[string]any{
			"loading": false,
		})
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

	GoWithHandlePanic(func() {
		s.runOpenAIStream(
			threadID, assistantMsg.ID, oaiMessages, isNewThread, sig.Message)
	})

	// Relay stream to this SSE connection
	s.relayStream(ctx, sse, stream, assistantDivID)

	// Signal loading done
	sse.MarshalAndPatchSignals(map[string]any{
		"loading": false,
	})
}

func (s *Server) runOpenAIStream(
	threadID uuid.UUID, assistantMsgID uuid.UUID,
	oaiMessages []openai.ChatCompletionMessageParamUnion,
	isNewThread bool, userMessage string,
) {
	activeStream, ok := s.streams.Get(threadID)
	if !ok {
		return
	}

	ctx := context.Background()

	logger := slog.With("threadId", threadID)

	logger.Info("starting openai stream")
	agent := agent.NewAgent(logger)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	lastSaved := ""

	for token, err := range agent.Stream(ctx, oaiMessages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}
		activeStream.Append(token)

		// Throttled DB save
		select {
		case <-ticker.C:
			content, _ := activeStream.Snapshot()
			if content != lastSaved {
				if err := s.db.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
					ID:      assistantMsgID,
					Content: content,
				}); err != nil {
					logger.Error("throttled save", "err", err)
				}
				lastSaved = content
			}
		default:
		}
	}

	// Final DB save
	finalContent, _ := activeStream.Snapshot()
	logger.Info("stream complete", "threadId", threadID, "contentLen", len(finalContent))
	if err := s.db.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
		ID:      assistantMsgID,
		Content: finalContent,
	}); err != nil {
		logger.Error("update assistant message", "err", err)
	}

	activeStream.Finish()

	// Keep stream in registry briefly so page refresh can resume.
	// Clean up after 30s if no one picks it up.
	go func() {
		time.Sleep(30 * time.Second)
		s.streams.DeleteIfSame(threadID, activeStream)
	}()

	// Auto-generate title for new threads
	if isNewThread {
		title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

		err := s.db.UpdateThreadTitle(ctx, db.UpdateThreadTitleParams{
			ID:    threadID,
			Title: title,
		})
		if err != nil {
			logger.Error("update thread title", "err", err)
		}
	}
}

func (s *Server) relayStream(ctx context.Context, sse *datastar.ServerSentEventGenerator, stream *ActiveStream, assistantDivID string) {
	ch, content, done := stream.Subscribe()
	defer stream.Unsubscribe(ch)

	if content != "" {
		assistantHTML := views.RenderAssistantMessage(stream.messageID, renderMarkdown(content), false)

		sse.PatchElements(
			assistantHTML,
			datastar.WithSelectorID(assistantDivID),
			datastar.WithModeOuter(),
		)
	}
	if done {
		return
	}

	for {
		select {
		case <-ch:
			content, done = stream.Snapshot()
			assistantHTML := views.RenderAssistantMessage(stream.messageID, renderMarkdown(content), false)

			sse.PatchElements(
				assistantHTML,
				datastar.WithSelectorID(assistantDivID),
				datastar.WithModeOuter(),
			)
			if done {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) handleResumeStream(w http.ResponseWriter, r *http.Request, threadID uuid.UUID) {
	stream, ok := s.streams.Get(threadID)
	if !ok {
		// Handle short race where refresh arrives slightly before stream registration.
		time.Sleep(150 * time.Millisecond)
		stream, ok = s.streams.Get(threadID)
	}
	if !ok {
		slog.Info("resume: no active stream", "threadId", threadID)
		sse := datastar.NewSSE(w, r)
		sse.MarshalAndPatchSignals(map[string]any{
			"loading": false,
		})
		return
	}

	slog.Info("resume: relaying active stream", "threadId", threadID, "messageId", stream.messageID)

	sse := datastar.NewSSE(w, r)
	assistantDivID := fmt.Sprintf("msg-%s", stream.messageID)

	s.relayStream(r.Context(), sse, stream, assistantDivID)

	slog.Info("resume: relay done", "threadId", threadID)
	sse.MarshalAndPatchSignals(map[string]any{
		"loading": false,
	})
}
