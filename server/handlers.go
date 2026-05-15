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

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (x *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	x.renderer.RenderIndexPage(w, views.IndexData{
		Threads:   mapViewThreads(threads),
		UserEmail: user.Email,
	})
}

func (x *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	balance, err := x.db.GetBillingBalanceCents(ctx, user.ID)
	if err != nil {
		handleBadReq(w, "get billing balance", err)
		return
	}

	x.renderer.RenderSettingsPage(w, views.SettingsData{
		Threads:      mapViewThreads(threads),
		UserEmail:    user.Email,
		Paid:         user.PaidAt != nil,
		BalanceCents: balance,
	})
}

func (x *Server) getDisplayMessages(
	ctx context.Context, userID, threadID string, withLargeHeight bool,
) ([]views.MessageGroup, error) {
	messages, err := x.db.GetMessagesByThread(ctx, userID, threadID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return mapViewMessages(messages, withLargeHeight), nil
}

func (x *Server) getThreadPage(
	ctx context.Context, user db.User, threadID string,
) (views.ThreadData, error) {
	var td views.ThreadData

	thread, err := x.db.GetThread(ctx, threadID, user.ID)
	if err != nil {
		return td, err
	}

	messages, err := x.getDisplayMessages(ctx, user.ID, threadID, false)
	if err != nil {
		return td, err
	}

	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		return td, fmt.Errorf("list threads: %w", err)
	}

	return views.ThreadData{
		Threads:   mapViewThreads(threads),
		Thread:    viewThreadFromDbThread(thread),
		Messages:  messages,
		Streaming: thread.IsStreaming,
		UserEmail: user.Email,
	}, nil
}

func (x *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
	ctx := r.Context()
	threadPage, err := x.getThreadPage(ctx, user, threadID)
	if err != nil {
		handleBadReq(w, "get thread", err)
		return
	}

	x.renderer.RenderThreadPage(w, threadPage, true)
}

func (x *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
	if err := x.db.ArchiveThread(ctx, user.ID, threadID); err != nil {
		slog.Error("archive thread", "err", err)
		toastShowError(w, r, "Unable to archive chat", "We couldn't archive this chat. Try again.")
		return
	}

	sse := datastar.NewSSE(w, r)

	sse.RemoveElementByID(fmt.Sprintf("thread-%s", threadID))

	// TODO: we should probably send the threadID from the path, so that we can just
	// check serverside whether to redirect instead of doing it clientside.
	sse.ExecuteScript(fmt.Sprintf(
		`if (window.location.pathname === '/threads/%s') window.location.replace('/')`,
		threadID,
	))
}

func (x *Server) handleSearchThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	var sig struct {
		ThreadSearch string `json:"threadSearch"`
		ThreadID     string `json:"threadId"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read search signals", "err", err)
		toastShowError(w, r, "Unable to search chats", "Refresh the page, then try searching again.")
		return
	}
	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		slog.Error("list threads for search", "err", err)
		toastShowError(w, r, "Unable to search chats", "We couldn't search your chats. Try again.")
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

func (x *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	var sig struct {
		Message  string  `json:"message"`
		ThreadID *string `json:"threadId"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read send message signals", "err", err)
		toastShowError(w, r, "Unable to send message", "Refresh the page, then try again.")
		return
	}
	balance, err := x.db.GetBillingBalanceCents(ctx, user.ID)
	if err != nil {
		slog.Error("get billing balance before send", "err", err)
		toastShowError(w, r, "Unable to send message", "Try again in a moment.")
		return
	}
	if balance <= 0 {
		toastShow(w, r, views.ToastData{
			Title:   "Usage balance empty",
			Message: "Add credit in Settings to keep chatting.",
			Action:  &views.ToastAction{Label: "Add credit", Href: "/settings"},
		})
		return
	}

	userMessage := sig.Message

	var threadID string
	isNewThread := sig.ThreadID == nil
	if isNewThread {
		id, err := x.db.InitThread(ctx, user.ID, "New Chat", "You are a helpful assistant")
		if err != nil {
			slog.Error("create thread before send", "err", err)
			toastShowError(w, r, "Unable to send message", "Try again.")
			return
		}

		threadID = id

		slog.Info("created thread", "threadId", threadID)

		GoRecover(func() {
			x.updateThreadTitle(context.Background(), user.ID, threadID, userMessage)
		})

		sse := datastar.NewSSE(w, r)

		slog.Info("redirecting")
		sse.Redirectf("/threads/%s", threadID)
	} else {
		_, err := x.db.GetThread(ctx, *sig.ThreadID, user.ID)
		if errors.Is(err, db.ErrNotFound) {
			toastShowError(w, r, "Chat not found", "This chat may have been removed. Go home and choose another chat.")
			return
		} else if err != nil {
			slog.Error("get thread before send", "err", err)
			toastShowError(w, r, "Unable to send message", "Try again.")
			return
		}

		threadID = *sig.ThreadID
	}

	GoRecover(func() {
		x.writeAgentStreamToThread(user.ID, threadID, userMessage)
	})
}

func (x *Server) persistStreamMessage(
	ctx context.Context, userID string, messages []db.Message, message db.Message,
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

	created, err := x.db.CreateMessageWithMeta(ctx, userID, message)
	if err != nil {
		return messages, err
	}

	return append(messages, created), nil
}

func (x *Server) writeAgentStreamToThread(userID, threadID, userMessage string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID, "userId", userID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	unsub := subscribeToStops(x.bus, threadID, cancel)
	defer unsub()

	if err := x.db.SetThreadStreaming(ctx, userID, threadID, true); err != nil {
		logger.Error("set streaming state", "err", err)
		return
	}

	x.publishSidebarEvent(sidebarEvent{threadID: threadID})
	defer func() {
		ctx := context.Background()
		if err := x.db.SetThreadStreaming(ctx, userID, threadID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
			return
		}
		x.publishThreadDidUpdateEventEvent(threadID)
		x.publishSidebarEvent(sidebarEvent{threadID: threadID})
	}()

	_, err := x.db.CreateUserMessage(ctx, userID, threadID, userMessage)
	if err != nil {
		logger.Error("create user message", "err", err)
		return
	}
	_, err = x.db.CreateAssistantMessage(ctx, userID, threadID, "")
	if err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}

	x.publishThreadEvent(threadID, ThreadEvent{UserMessageCreated: true})

	messages, err := x.db.GetMessagesByThread(ctx, userID, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	agent := agents.NewAgent(x.env)

	for tokenEvt, err := range agent.Stream(ctx, messages) {
		if err != nil {
			logger.Error("stream error", "err", err)
			break
		}

		switch evt := tokenEvt.(type) {
		case agents.AssistantDelta:
			messages, err = x.persistStreamMessage(ctx, userID, messages, db.Message{
				ThreadID: threadID,
				Type:     "assistant_message",
				Content:  evt.Delta,
			})
		case agents.ReasoningDelta:
			messages, err = x.persistStreamMessage(ctx, userID, messages, db.Message{
				ThreadID: threadID,
				Type:     "assistant_reasoning",
				Content:  evt.Delta,
			})
		case agents.Toolcall:
			messages, err = x.persistStreamMessage(ctx, userID, messages, db.Message{
				ThreadID:   threadID,
				Type:       "assistant_tool_call",
				Content:    evt.Args,
				ToolCallID: evt.CallID,
				ToolName:   evt.Name,
			})
		case agents.ToolOutput:
			messages, err = x.persistStreamMessage(ctx, userID, messages, db.Message{
				ThreadID:   threadID,
				Type:       "tool_output",
				Content:    evt.Output,
				ToolCallID: evt.CallID,
				ToolName:   evt.Name,
			})
		case agents.UsageDelta:
			logger.Debug("received usage",
				"provider", evt.Usage.Provider,
				"model", evt.Usage.Model,
				"input_tokens", evt.Usage.InputTokens,
				"cached_input_tokens", evt.Usage.CachedInputTokens,
				"output_tokens", evt.Usage.OutputTokens,
				"reasoning_tokens", evt.Usage.ReasoningTokens,
			)
			_, err = x.db.CreateUsage(ctx, db.Usage{
				UserID:            userID,
				ThreadID:          threadID,
				MessageID:         lastAssistantMessageID(messages),
				Provider:          evt.Usage.Provider,
				Model:             evt.Usage.Model,
				InputTokens:       evt.Usage.InputTokens,
				CachedInputTokens: evt.Usage.CachedInputTokens,
				OutputTokens:      evt.Usage.OutputTokens,
				ReasoningTokens:   evt.Usage.ReasoningTokens,
			})
		}
		if err != nil {
			logger.Error("persist stream message", "err", err)
			break
		}

		if _, ok := tokenEvt.(agents.UsageDelta); !ok {
			x.publishThreadDidUpdateEventEvent(threadID)
		}
	}
}

func lastAssistantMessageID(messages []db.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Type == db.AssistantMessage {
			return messages[i].ID
		}
	}
	return ""
}

func (x *Server) updateThreadTitle(ctx context.Context, userID, threadID string, userMessage string) {
	agent := agents.NewAgent(x.env)
	title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

	err := x.db.UpdateThreadTitle(ctx, userID, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}

	x.publishSidebarEvent(sidebarEvent{threadID: threadID})
}

func (x *Server) handleThreadEvents(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
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
		threadPage, err := x.getThreadPage(ctx, user, threadID)
		if err != nil {
			slog.Error("load thread event page", "err", err)
			views.PatchErrorToast(sse, "Unable to load updates", "Refresh the page to reconnect this chat.")
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

			thread, err := x.db.GetThread(ctx, threadID, user.ID)
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			messages, err := x.getDisplayMessages(ctx, user.ID, threadID, true)
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

			thread, err := x.db.GetThread(ctx, evt.threadID, user.ID)
			if errors.Is(err, db.ErrNotFound) {
				continue msgloop
			}
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			views.PatchThreadItem(sse, viewThreadFromDbThread(thread))
		case <-ctx.Done():
			break msgloop
		}
	}
}

func (x *Server) handleStopStreaming(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	var sig struct {
		ThreadID string `json:"threadId"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read stop streaming signals", "err", err)
		toastShowError(w, r, "Unable to stop response", "Refresh the page, then try again.")
		return
	}
	if _, err := x.db.GetThread(ctx, sig.ThreadID, user.ID); err != nil {
		slog.Error("get thread before stop", "err", err)
		toastShowError(w, r, "Unable to stop response", "We couldn't stop this response. Try again.")
		return
	}

	publishThreadStop(x.bus, sig.ThreadID)
}
