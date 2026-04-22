package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"afastchat.com/agents"
	"afastchat.com/db"
	"afastchat.com/views"

	"github.com/starfederation/datastar-go/datastar"
)

func (x *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	x.renderer.RenderLandingPage(w)
}

func (x *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	x.renderer.RenderLoginPage(w)
}

func (x *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	threads, err := x.db.ListThreads(r.Context())
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	x.renderer.RenderIndexPage(w, views.IndexData{Threads: mapViewThreads(threads)}, true)
}

func (x *Server) getDisplayMessages(
	ctx context.Context, threadID string, withLargeHeight bool) ([]views.MessageGroup, error) {
	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return mapViewMessages(messages, withLargeHeight), nil
}

func (x *Server) getThreadPage(ctx context.Context, threadID string) (views.ThreadData, error) {
	var td views.ThreadData

	thread, err := x.db.GetThread(ctx, threadID)
	if err != nil {
		return td, err
	}

	messages, err := x.getDisplayMessages(ctx, threadID, false)
	if err != nil {
		return td, err
	}

	threads, err := x.db.ListThreads(ctx)
	if err != nil {
		return td, fmt.Errorf("list threads: %w", err)
	}

	return views.ThreadData{
		Threads:   mapViewThreads(threads),
		Thread:    mapViewThread(thread),
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

	x.renderer.RenderThreadPage(w, threadPage, true)
}

func (x *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request, threadID string) {
	if err := x.db.ArchiveThread(r.Context(), threadID); err != nil {
		handleBadReq(w, "archive thread", err)
		return
	}

	sse := datastar.NewSSE(w, r)

	sse.RemoveElementByID(fmt.Sprintf("thread-%s", threadID))
	sse.ExecuteScript(fmt.Sprintf(
		`if (window.location.pathname === '/threads/%s') window.location.replace('/app')`,
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

	query := sig.ThreadSearch
	filtered := make([]db.Thread, 0, len(threads))
	for _, it := range threads {
		if strings.Contains(strings.ToLower(it.Title), query) {
			filtered = append(filtered, it)
		}
	}

	sse := datastar.NewSSE(w, r)
	views.PatchThreads(sse, mapViewThreads(filtered), sig.ThreadID)
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

		GoRecover(func() {
			x.updateThreadTitle(context.Background(), threadID, userMessage)
		})

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

	GoRecover(func() {
		x.writeAgentStreamToThread(threadID, userMessage)
	})
}

func (x *Server) persistStreamMessage(
	ctx context.Context, messages []db.Message, message db.Message,
) ([]db.Message, error) {
	if len(messages) > 0 {
		last := &messages[len(messages)-1]
		if message.Type == "assistant_message" || message.Type == "assistant_reasoning" {
			if last.Type == message.Type {
				last.Content += message.Content
				err := x.db.UpdateMessageContent(ctx, last.ID, last.Content)
				return messages, err
			}
		}
	}

	created, err := x.db.CreateMessageWithMeta(ctx, message)
	if err != nil {
		return messages, err
	}

	return append(messages, created), nil
}

func (x *Server) writeAgentStreamToThread(threadID, userMessage string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	unsub := subscribeToStops(x.bus, threadID, cancel)
	defer unsub()

	if err := x.db.SetThreadStreaming(ctx, threadID, true); err != nil {
		logger.Error("set streaming state", "err", err)
		return
	}

	x.publishSidebarEvent(sidebarEvent{threadID: threadID})
	defer func() {
		ctx := context.Background()
		if err := x.db.SetThreadStreaming(ctx, threadID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
			return
		}
		x.publishThreadDidUpdateEventEvent(threadID)
		x.publishSidebarEvent(sidebarEvent{threadID: threadID})
	}()

	_, err := x.db.CreateUserMessage(ctx, threadID, userMessage)
	if err != nil {
		logger.Error("create user message", "err", err)
		return
	}
	_, err = x.db.CreateAssistantMessage(ctx, threadID, "")
	if err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}

	x.publishThreadEvent(threadID, ThreadEvent{UserMessageCreated: true})

	messages, err := x.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	agent := agents.NewAgent()

	for tokenEvt, err := range agent.Stream(ctx, messages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		switch evt := tokenEvt.(type) {
		case agents.AssistantDelta:
			messages, err = x.persistStreamMessage(ctx, messages, db.Message{
				ThreadID: threadID,
				Type:     "assistant_message",
				Content:  evt.Delta,
			})
		case agents.ReasoningDelta:
			messages, err = x.persistStreamMessage(ctx, messages, db.Message{
				ThreadID: threadID,
				Type:     "assistant_reasoning",
				Content:  evt.Delta,
			})
		case agents.Toolcall:
			messages, err = x.persistStreamMessage(ctx, messages, db.Message{
				ThreadID:   threadID,
				Type:       "assistant_tool_call",
				Content:    evt.Args,
				ToolCallID: evt.CallID,
				ToolName:   evt.Name,
			})
		case agents.ToolOutput:
			messages, err = x.persistStreamMessage(ctx, messages, db.Message{
				ThreadID:   threadID,
				Type:       "tool_output",
				Content:    evt.Output,
				ToolCallID: evt.CallID,
				ToolName:   evt.Name,
			})
		}
		if err != nil {
			logger.Error("persist stream message", "err", err)
			break
		}

		x.publishThreadDidUpdateEventEvent(threadID)
	}
}

func (x *Server) updateThreadTitle(ctx context.Context, threadID string, userMessage string) {
	title := agents.GenerateTitleFromUserMessage(ctx, userMessage)

	err := x.db.UpdateThreadTitle(ctx, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}

	x.publishSidebarEvent(sidebarEvent{threadID: threadID})
}

func (x *Server) handleThreadEvents(w http.ResponseWriter, r *http.Request, threadID string) {
	ctx := r.Context()

	threadEvtsC := make(chan ThreadEvent, 100)
	threadEventHandler := func(evt ThreadEvent) {
		threadEvtsC <- evt
	}

	unsub := x.subscribeToThread(threadID, threadEventHandler)
	defer unsub()

	sidebarEvtsC := make(chan sidebarEvent, 100)
	sidebarHandler := func(evt sidebarEvent) {
		sidebarEvtsC <- evt
	}

	unsub = x.subscribeToSidebar(sidebarHandler)
	defer unsub()

	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli()))
	{
		threadPage, err := x.getThreadPage(ctx, threadID)
		if err != nil {
			handleBadReq(w, "get thread", err)
			return
		}

		var sb strings.Builder

		x.renderer.RenderThreadPage(&sb, threadPage, false)
		sse.PatchElements(sb.String())
	}

msgloop:
	for {
		select {
		case evt := <-threadEvtsC:
			slog.Debug("received thread event")

			thread, err := x.db.GetThread(ctx, threadID)
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			messages, err := x.getDisplayMessages(ctx, threadID, true)
			if err != nil {
				slog.Error("error getting messages", "err", err)
				continue msgloop
			}

			if evt.UserMessageCreated {
				views.PatchThreadMessageList(sse, messages, true)
				views.ClearMessageInput(sse)
			} else {
				views.PatchThreadMessageList(sse, messages, false)
				views.PatchIsStreaming(sse, thread.IsStreaming)
			}

		case evt := <-sidebarEvtsC:
			slog.Debug("received sidebar thread update")

			thread, err := x.db.GetThread(ctx, evt.threadID)
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			views.PatchThreadItem(sse, mapViewThread(thread))
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
