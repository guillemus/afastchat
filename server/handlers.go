package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"afastchat.com/agents"
	"afastchat.com/db"
	"afastchat.com/views"

	"github.com/starfederation/datastar-go/datastar"
)

func (x *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	threads, err := x.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	x.renderer.RenderIndex(w, views.IndexData{Threads: threads})
}

func (x *Server) getThreadPage(ctx context.Context, threadID string) (*views.ThreadData, error) {
	thread, err := x.db.GetThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	for i := range messages {
		if messages[i].Role == "assistant" {
			messages[i].Content = renderMarkdown(messages[i].Content)
		}
	}

	threads, err := x.db.ListThreads(ctx)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}

	return &views.ThreadData{
		Threads:   threads,
		Thread:    thread,
		Messages:  messages,
		Streaming: thread.IsStreaming,
	}, nil
}

func (x *Server) handleThread(w http.ResponseWriter, r *http.Request, threadID string) {
	ctx := r.Context()

	threadPage, err := x.getThreadPage(ctx, threadID)
	if err != nil {
		handleBadReq(w, "get thread", err)
		return
	}

	x.renderer.RenderThread(w, *threadPage)
}

func (x *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request, threadID string) {
	if err := x.db.ArchiveThread(r.Context(), threadID); err != nil {
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

func (x *Server) handleSearchThreads(w http.ResponseWriter, r *http.Request, sig SearchThreadsSig) {
	threads, err := x.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	threads, err = x.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	query := sig.ThreadSearch
	filtered := make([]db.Thread, 0, len(threads))
	for _, it := range threads {
		if strings.Contains(strings.ToLower(it.Title), query) {
			filtered = append(filtered, it)
		}
	}

	var threadsSb strings.Builder
	views.RenderThreads(&threadsSb, filtered, sig.ThreadID)

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(
		threadsSb.String(),
		datastar.WithSelectorID("thread-items"),
	)
}

type SendMessageSignals struct {
	Message  string  `json:"message"`
	ThreadID *string `json:"threadId"`
}

func (x *Server) handleSendMessage(w http.ResponseWriter, r *http.Request, sig SendMessageSignals) {
	ctx := r.Context()

	userMessage := sig.Message

	var threadID string
	isNewThread := sig.ThreadID == nil
	if isNewThread {
		id, err := x.db.InitThread(ctx, "New Chat", "You are a helpful assistant", userMessage)
		if err != nil {
			handleBadReq(w, "create thread", err)
			return
		}

		threadID = id

		slog.Info("created thread", "threadId", threadID)

		go func() {
			defer recoverAndLog()
			x.updateThreadTitle(context.Background(), threadID, userMessage)
		}()

		sse := datastar.NewSSE(w, r)

		slog.Info("redirecting")
		sse.Redirectf("/threads/%s", threadID)
	} else {
		_, err := x.db.GetThread(ctx, *sig.ThreadID)
		if errors.Is(err, db.ErrNotFound) {
			handleBadReq(w, "thread not found", err)
			return
		} else if err != nil {
			handleBadReq(w, "get thread", err)
			return
		}

		threadID = *sig.ThreadID

		dbUserMessage, err := x.db.CreateUserMessage(ctx, threadID, userMessage)
		if err != nil {
			handleBadReq(w, "create user message", err)
			return
		}

		publishThreadMessage(x.bus, threadID, threadEvent{
			isFromUser: true,
			messageID:  dbUserMessage.ID,
			message:    dbUserMessage.Content,
		})
	}

	sse := datastar.NewSSE(w, r)
	patchSignals(sse, map[string]any{"message": ""})

	x.writeAgentStreamToThread(threadID)
}

func (x *Server) writeAgentStreamToThread(threadID string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	x.agentCanceller.set(threadID, cancel)
	defer x.agentCanceller.clear(threadID)

	// Set streaming state
	if err := x.db.SetThreadStreaming(ctx, threadID, true); err != nil {
		logger.Error("set streaming state", "err", err)
		return
	}
	defer func() {
		if err := x.db.SetThreadStreaming(ctx, threadID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
		}
	}()

	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	agent := agents.NewAgent()

	// create empty assistant message so that we have an id to go and update.
	assistantMsgID, err := x.db.CreateAssistantMessage(ctx, threadID, "")
	if err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}

	publishThreadMessage(x.bus, threadID, threadEvent{
		isFirst:   true,
		messageID: assistantMsgID,
		message:   "",
	})

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var assistantMsg strings.Builder

	for token, err := range agent.Stream(ctx, messages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		assistantMsg.WriteString(token)

		publishThreadMessage(x.bus, threadID, threadEvent{
			messageID: assistantMsgID,
			message:   assistantMsg.String(),
		})

		select {
		case <-ticker.C:
			logger.Debug("updating assistant message")
			err := x.db.UpdateMessageContent(ctx, assistantMsgID, assistantMsg.String())
			if err != nil {
				logger.Error("update assistant message", "err", err)
			}
		default:
		}
	}

	// By now, context might be cancelled already because of of a thread stop. We should use a fresh context.
	ctx = context.Background()

	logger.Debug("updating assistant message")
	err = x.db.UpdateMessageContent(ctx, assistantMsgID, assistantMsg.String())
	if err != nil {
		logger.Error("final update assistant message", "err", err)
	}
}

func (x *Server) updateThreadTitle(ctx context.Context, threadID string, userMessage string) {
	agent := agents.NewAgent()
	title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

	err := x.db.UpdateThreadTitle(ctx, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}
}

func (x *Server) handleThreadEvents(w http.ResponseWriter, r *http.Request, threadID string) {
	ctx := r.Context()

	evtsC := make(chan threadEvent, 100)
	handler := func(evt threadEvent) {
		evtsC <- evt
	}

	unsub := subscribeToThread(x.bus, threadID, handler)
	defer unsub()

	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli()))
	threadPage, err := x.getThreadPage(ctx, threadID)
	if err != nil {
		handleBadReq(w, "get thread", err)
		return
	}

	var sb strings.Builder
	x.renderer.RenderThread(&sb, *threadPage)

	sse.PatchElements(sb.String())

msgloop:
	for {
		select {
		case evt := <-evtsC:
			slog.Debug("received event", "evt", evt)

			if evt.isFromUser {
				var sb strings.Builder
				views.RenderUserMessage(&sb, evt.messageID, evt.message)
				sse.PatchElements(sb.String(),
					datastar.WithSelectorID("messages"),
					datastar.WithModeAppend(),
				)

				continue
			}

			html := renderMarkdown(evt.message)

			var sb strings.Builder
			views.RenderAssistantMessage(&sb, evt.messageID, html)

			if evt.isFirst {
				sse.PatchElements(sb.String(),
					datastar.WithSelectorID("messages"),
					datastar.WithModeAppend(),
				)
			} else {
				sse.PatchElements(sb.String())
			}

		case <-ctx.Done():
			break msgloop
		}
	}
}

type StopStreamingSignals struct {
	ThreadID string `json:"threadId"`
}

func (x *Server) handleStopStreaming(
	w http.ResponseWriter, r *http.Request, sig StopStreamingSignals) {

	cancel, ok := x.agentCanceller.get(sig.ThreadID)
	if ok {
		cancel()
	}
}
