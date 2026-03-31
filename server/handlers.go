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

	userMessage := sig.Message

	var threadID uuid.UUID
	isNewThread := sig.ThreadID == ""
	if isNewThread {
		id, err := s.db.CreateThread(ctx, "New Chat", "", userMessage)
		if err != nil {
			handleBadReq(w, "create thread", err)
			return
		}
		slog.Info("created thread", "threadId", threadID)
		threadID = *id
	} else {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			handleBadReq(w, "invalid thread id", err)
			return
		}

		_, err = s.db.GetThread(ctx, threadID)
		if errors.Is(err, sql.ErrNoRows) {
			handleBadReq(w, "thread not found", err)
			return
		} else if err != nil {
			handleBadReq(w, "get thread", err)
			return
		}

		threadID = parsed
	}

	if isNewThread {
		go func() {
			defer recoverAndLog()
			s.updateThreadTitle(ctx, threadID, userMessage)
		}()

		sse := datastar.NewSSE(w, r)
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

	messages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		// fixme: log err
		return
	}

	logger.Info("starting agent stream")
	agent := agents.NewAgent()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for token, err := range agent.Stream(ctx, messages) {
		_ = token // we will fix this alter
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		// Throttled DB save
		select {
		case <-ticker.C:

		default:
		}
	}

	// fixme: final db save
}

func (s *Server) updateThreadTitle(ctx context.Context, threadID uuid.UUID, userMessage string) {
	agent := agents.NewAgent()
	title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

	err := s.db.UpdateThreadTitle(ctx, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}
}
