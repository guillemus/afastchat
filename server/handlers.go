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

func (x *Server) getThreadPage(ctx context.Context, threadID string) (views.ThreadData, error) {
	var td views.ThreadData

	thread, err := x.db.GetThread(ctx, threadID)
	if err != nil {
		return td, err
	}

	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		return td, fmt.Errorf("get messages: %w", err)
	}

	for i := range messages {
		if messages[i].Role == "assistant" {
			messages[i].Content = renderMarkdown(messages[i].Content)
		}
	}

	threads, err := x.db.ListThreads(ctx)
	if err != nil {
		return td, fmt.Errorf("list threads: %w", err)
	}

	return views.ThreadData{
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

	if isDatastarRequest(r) {
		x.renderer.RenderThread(w, threadPage, false)
	} else {
		x.renderer.RenderThread(w, threadPage, true)
	}
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
		id, err := x.db.InitThread(ctx, "New Chat", "You are a helpful assistant")
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
	}

	sse := datastar.NewSSE(w, r)
	patchSignals(sse, map[string]any{"message": ""})

	go func() {
		defer recoverAndLog()
		x.writeAgentStreamToThread(threadID, userMessage)
	}()
}

func (x *Server) writeAgentStreamToThread(threadID, userMessage string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	unsub := subscribeToStops(x.bus, threadID, cancel)
	defer unsub()

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

	dbUserMessage, err := x.db.CreateUserMessage(ctx, threadID, userMessage)
	if err != nil {
		logger.Error("create user message", "err", err)
		return
	}

	// create empty assistant message so that we have an id to go and update.
	assistantMsgID, err := x.db.CreateAssistantMessage(ctx, threadID, "")
	if err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}

	x.publishThreadEvent(threadID, evtAssistantStart{
		userMessageID:      dbUserMessage.ID,
		userMessage:        userMessage,
		assistantMessageID: assistantMsgID,
	})

	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	agent := agents.NewAgent()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var assistantMsg strings.Builder

	for token, err := range agent.Stream(ctx, messages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		assistantMsg.WriteString(token)

		x.publishThreadEvent(threadID, evtAssistantChunk{
			messageID: assistantMsgID,
			content:   assistantMsg.String(),
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

	x.publishThreadEvent(threadID, evtFinishedStreaming{})

	// By now, context might be cancelled already because of of a thread stop.
	// We should use a fresh context.
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

	x.publishThreadEvent(threadID, evtTitleUpdate{newTitle: title})
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
	x.renderer.RenderThread(&sb, threadPage, false)

	sse.PatchElements(sb.String())

msgloop:
	for {
		select {
		case event := <-evtsC:
			slog.Debug("received event", "evt", event)

			switch evt := event.(type) {
			case evtAssistantStart:

				patchSignals(sse, map[string]any{"isStreaming": true})

				sse.ExecuteScript(`(() => {
					const el = document.querySelector('#messages .assistant-message.last')
					if (el) {
						el.classList.remove('last')
					}
				})()`)

				var sb strings.Builder
				views.RenderUserMessage(&sb, evt.userMessageID, evt.userMessage)
				views.RenderAssistantMessageAsLast(&sb, evt.assistantMessageID, "")

				sse.PatchElements(sb.String(),
					datastar.WithSelectorID("messages"),
					datastar.WithModeAppend(),
				)

				sse.ExecuteScript(`(() => {
				  	const el = document.getElementById('messages')
				  	if (el) {
						el.scrollTop = el.scrollHeight
					}
				})()`)

			case evtAssistantChunk:

				html := renderMarkdown(evt.content)

				var sb strings.Builder
				views.RenderAssistantMessageAsLast(&sb, evt.messageID, html)
				sse.PatchElements(sb.String())

			case evtFinishedStreaming:
				patchSignals(sse, map[string]any{"isStreaming": false})

			case evtTitleUpdate:
				thread, err := x.db.GetThread(ctx, threadID)
				if err != nil {
					slog.Error("error getting thread", "err", err)
					continue msgloop
				}

				var sb strings.Builder
				views.RenderThreadItem(&sb, thread)
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

	publishThreadStop(x.bus, sig.ThreadID)
}
