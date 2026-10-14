package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"afastchat.com/agents"
	"afastchat.com/db"
	"afastchat.com/views"
	"afastchat.com/views/ui"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
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

	models, err := x.modelPicker(ctx, user)
	if err != nil {
		handleBadReq(w, "load model picker", err)
		return
	}

	x.render(w, r, views.IndexPage(views.IndexData{
		Threads: mapViewThreads(threads),
		Viewer:  viewerFromUser(user),
		Models:  models,
	}))
}

func (x *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		handleBadReq(w, "list threads", err)
		return
	}

	_, connected, err := x.openRouterKey(ctx, user.ID)
	if err != nil {
		handleBadReq(w, "load openrouter key", err)
		return
	}

	x.render(w, r, views.SettingsPage(views.SettingsData{
		Threads:             mapViewThreads(threads),
		Viewer:              viewerFromUser(user),
		OpenRouterConnected: connected,
		OpenRouterFailed:    r.URL.Query().Get("openrouter_failed") == "1",
	}))
}

func (x *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)

	streaming, err := x.db.ListStreamingThreads(ctx)
	if err != nil {
		slog.Error("list streaming threads before account deletion", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Your account could not be deleted. Try again."})
		return
	}
	for _, thread := range streaming {
		if thread.UserID == user.ID {
			x.broker.PublishThreadStop(thread.ID)
		}
	}

	if err := x.db.DeleteUser(ctx, user.ID); err != nil {
		slog.Error("delete account", "user_id", user.ID, "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Your account could not be deleted. Try again."})
		return
	}

	if err := x.sessions.Destroy(ctx); err != nil {
		slog.Error("delete account: destroy session", "err", err)
	}
	slog.Info("account deleted", "user_id", user.ID)

	sse := datastar.NewSSE(w, r)
	if err := sse.Redirect("/"); err != nil {
		slog.Error("redirect after account deletion", "err", err)
	}
}

func (x *Server) getDisplayMessages(
	ctx context.Context, userID, threadID string, largeHeight, streaming bool,
) ([]views.MessageGroup, error) {
	messages, err := x.db.GetMessagesByThread(ctx, userID, threadID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return mapViewMessages(messages, streaming, largeHeight), nil
}

func (x *Server) getThreadPage(
	ctx context.Context, user db.User, threadID string,
) (views.ThreadData, error) {
	var td views.ThreadData

	thread, err := x.db.GetThread(ctx, threadID, user.ID)
	if err != nil {
		return td, err
	}

	messages, err := x.getDisplayMessages(ctx, user.ID, threadID, thread.IsStreaming, thread.IsStreaming)
	if err != nil {
		return td, err
	}

	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		return td, fmt.Errorf("list threads: %w", err)
	}

	models, err := x.modelPicker(ctx, user)
	if err != nil {
		return td, fmt.Errorf("load model picker: %w", err)
	}

	return views.ThreadData{
		Threads:   mapViewThreads(threads),
		Thread:    viewThreadFromDbThread(thread),
		Messages:  messages,
		Streaming: thread.IsStreaming,
		Viewer:    viewerFromUser(user),
		Models:    models,
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

	x.render(w, r, views.ThreadPage(threadPage, true))
}

func (x *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
	if err := x.db.ArchiveThread(ctx, user.ID, threadID); err != nil {
		slog.Error("archive thread", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "We couldn't archive this chat. Try again."})
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
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to search chats. Refresh the page, then try again."})
		return
	}
	threads, err := x.db.ListThreads(ctx, user.ID)
	if err != nil {
		slog.Error("list threads for search", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "We couldn't search your chats. Try again."})
		return
	}

	query := strings.ToLower(strings.TrimSpace(sig.ThreadSearch))
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
		Message  string `json:"message"`
		ThreadID string `json:"threadId"`
		Model    string `json:"model"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read send message signals", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to send message. Refresh the page, then try again."})
		return
	}
	userMessage := sig.Message
	apiKey := mustGetOpenRouterKey(r)
	modelID := agents.ResolveModel(sig.Model).ID
	if modelID != user.SelectedModel {
		if err := x.db.UpdateUserSelectedModel(ctx, user.ID, modelID); err != nil {
			slog.Error("save selected model before send", "err", err)
		}
	}

	var threadID string
	isNewThread := sig.ThreadID == ""
	if isNewThread {
		id, err := x.db.InitThread(ctx, user.ID, "New Chat", "You are a helpful assistant")
		if err != nil {
			slog.Error("create thread before send", "err", err)
			x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to send message. Try again."})
			return
		}

		threadID = id

		slog.Info("created thread", "threadId", threadID)

		GoRecover(func() {
			x.updateThreadTitle(context.Background(), apiKey, user.ID, threadID, userMessage)
		})

		sse := datastar.NewSSE(w, r)

		slog.Info("redirecting")
		sse.Redirectf("/threads/%s", threadID)
	} else {
		_, err := x.db.GetThread(ctx, sig.ThreadID, user.ID)
		if errors.Is(err, db.ErrNotFound) {
			x.patchToast(w, r, ui.Toast{Kind: ui.ToastWarning, Message: "Chat not found. It may have been removed."})
			return
		} else if err != nil {
			slog.Error("get thread before send", "err", err)
			x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to send message. Try again."})
			return
		}

		threadID = sig.ThreadID
	}

	GoRecover(func() {
		x.writeAgentStreamToThread(apiKey, modelID, user.ID, threadID, userMessage)
	})
}

func (x *Server) writeAgentStreamToThread(apiKey, modelID, userID, threadID, userMessage string) {
	ctx := context.Background()

	logger := slog.With("threadId", threadID, "userId", userID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := x.cancelOnThreadStop(ctx, threadID, cancel); err != nil {
		logger.Error("subscribe to thread stops", "err", err)
	}

	if err := x.db.SetThreadStreaming(ctx, userID, threadID, true); err != nil {
		logger.Error("set streaming state", "err", err)
		return
	}

	x.broker.PublishUserThread(userID, threadID)
	defer func() {
		ctx := context.Background()
		if err := x.db.SetThreadStreaming(ctx, userID, threadID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
			return
		}
		x.broker.PublishThreadUpdated(threadID)
		x.broker.PublishUserThread(userID, threadID)
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

	x.broker.PublishThread(threadID, threadEvent{UserMessageCreated: true})

	messages, err := x.db.GetMessagesByThread(ctx, userID, threadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	x.streamAgentMessages(ctx, apiKey, modelID, userID, threadID, messages, logger)
}

const streamFlushInterval = 50 * time.Millisecond

func (x *Server) streamAgentMessages(
	ctx context.Context, apiKey, modelID, userID, threadID string, messages []db.Message, logger *slog.Logger,
) {
	agent := agents.NewAgent(apiKey, modelID)
	w := &streamWriter{
		x:        x,
		ctx:      context.WithoutCancel(ctx),
		logger:   logger,
		userID:   userID,
		threadID: threadID,
		messages: slices.Clone(messages),
	}
	defer func() {
		if err := w.flush(); err != nil {
			logger.Error("flush stream message", "err", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	events := pullStream(ctx, agent.Stream(ctx, messages))

	ticker := time.NewTicker(streamFlushInterval)
	defer ticker.Stop()

	for {
		var err error
		select {
		case <-ticker.C:
			err = w.flush()
		case item, ok := <-events:
			if !ok {
				return
			}
			if item.err != nil {
				logger.Error("stream error", "err", item.err)
				return
			}
			err = w.write(item.evt)
		}
		if err != nil {
			logger.Error("persist stream message", "err", err)
			return
		}
	}
}

func lastAssistantMessageID(messages []db.Message) string {
	for _, message := range slices.Backward(messages) {
		if message.Type == db.AssistantMessage {
			return message.ID
		}
	}
	return ""
}

func (x *Server) updateThreadTitle(ctx context.Context, apiKey, userID, threadID string, userMessage string) {
	agent := agents.NewAgent(apiKey, "")
	title := agent.GenerateTitleFromUserMessage(ctx, userMessage)

	err := x.db.UpdateThreadTitle(ctx, userID, threadID, title)
	if err != nil {
		slog.Error("update thread title", "err", err)
	}

	x.broker.PublishUserThread(userID, threadID)
}

func (x *Server) handleThreadEvents(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
	ctx := r.Context()

	threadSubject := ThreadSubject(threadID)
	userThreadsSubject := UserThreadsSubject(user.ID)
	messagesC := make(chan *nats.Msg, 64)
	unsubscribe, err := x.broker.SubscribeSubjects(messagesC, threadSubject, userThreadsSubject)
	if err != nil {
		slog.Error("subscribe to thread events", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer unsubscribe()

	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli(
		datastar.WithBrotliLGWin(18),
		datastar.WithBrotliLevel(3),
	)))
	{
		threadPage, err := x.getThreadPage(ctx, user, threadID)
		if err != nil {
			slog.Error("load thread event page", "err", err)
			patchToastOnSSE(sse, ui.Toast{
				Kind:    ui.ToastError,
				Message: "Unable to load updates. Refresh the page to reconnect this chat.",
			})
			return
		}

		if err := sse.PatchElementTempl(views.ThreadPage(threadPage, false)); err != nil {
			slog.Error("patch thread page", "err", err)
			return
		}
	}

msgloop:
	for ctx.Err() == nil {
		var msg *nats.Msg
		select {
		case msg = <-messagesC:
		case <-ctx.Done():
			break msgloop
		}

		switch msg.Subject {
		case threadSubject:
			slog.Debug("received thread event")

			var evt threadEvent
			if err := json.Unmarshal(msg.Data, &evt); err != nil {
				slog.Error("decode thread event", "err", err)
				continue msgloop
			}

			thread, err := x.db.GetThread(ctx, threadID, user.ID)
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			messages, err := x.getDisplayMessages(ctx, user.ID, threadID, true, thread.IsStreaming)
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

		case userThreadsSubject:
			slog.Debug("received sidebar thread update")

			thread, err := x.db.GetThread(ctx, string(msg.Data), user.ID)
			if errors.Is(err, db.ErrNotFound) {
				continue msgloop
			}
			if err != nil {
				slog.Error("error getting thread", "err", err)
				continue msgloop
			}

			views.PatchThreadItem(sse, viewThreadFromDbThread(thread), thread.ID == threadID)
			if thread.ID == threadID {
				views.PatchThreadHeader(sse, viewThreadFromDbThread(thread))
			}
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
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to stop the response. Refresh the page, then try again."})
		return
	}
	if _, err := x.db.GetThread(ctx, sig.ThreadID, user.ID); err != nil {
		slog.Error("get thread before stop", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "We couldn't stop this response. Try again."})
		return
	}

	x.broker.PublishThreadStop(sig.ThreadID)
}

func viewerFromUser(user db.User) views.Viewer {
	return views.Viewer{Name: user.DisplayName, Email: user.Email}
}

func (x *Server) handleRenameThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")
	var sig struct {
		RenameTitle string `json:"renameTitle"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read rename signals", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Unable to rename this chat. Refresh the page, then try again."})
		return
	}

	title := strings.TrimSpace(sig.RenameTitle)
	if title == "" {
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastWarning, Message: "Enter a name for this chat."})
		return
	}

	if err := x.db.UpdateThreadTitle(ctx, user.ID, threadID, title); err != nil {
		slog.Error("rename thread", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "We couldn't rename this chat. Try again."})
		return
	}

	x.broker.PublishUserThread(user.ID, threadID)

	sse := datastar.NewSSE(w, r)
	if err := sse.ExecuteScript(`document.getElementById('rename-thread-dialog')?.close()`); err != nil {
		slog.Error("close rename dialog", "err", err)
	}
}

func (x *Server) handleDeleteThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	threadID := chi.URLParam(r, "id")

	x.broker.PublishThreadStop(threadID)
	if err := x.db.DeleteThread(ctx, user.ID, threadID); err != nil {
		slog.Error("delete thread", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "We couldn't delete this chat. Try again."})
		return
	}

	sse := datastar.NewSSE(w, r)
	if err := sse.Redirect("/"); err != nil {
		slog.Error("redirect after thread deletion", "err", err)
	}
}
