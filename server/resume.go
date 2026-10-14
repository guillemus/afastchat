package server

import (
	"context"
	"log/slog"

	"afastchat.com/db"
)

const resumeMessage = "Continue your previous response exactly where you left off. Do not mention that you are continuing."

func (x *Server) resumeStreamingThreads() {
	ctx := context.Background()
	threads, err := x.db.ListStreamingThreads(ctx)
	if err != nil {
		slog.Error("list streaming threads", "err", err)
		return
	}

	for _, thread := range threads {
		GoRecover(func() {
			x.resumeStreamingThread(thread)
		})
	}
}

func (x *Server) resumeStreamingThread(thread db.Thread) {
	ctx := context.Background()
	logger := slog.With("threadId", thread.ID, "userId", thread.UserID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	unsub := subscribeToStops(x.bus, thread.ID, cancel)
	defer unsub()

	x.publishSidebarEvent(sidebarEvent{threadID: thread.ID})
	defer func() {
		ctx := context.Background()
		if err := x.db.SetThreadStreaming(ctx, thread.UserID, thread.ID, false); err != nil {
			logger.Error("unset streaming state", "err", err)
			return
		}
		x.publishThreadDidUpdateEventEvent(thread.ID)
		x.publishSidebarEvent(sidebarEvent{threadID: thread.ID})
	}()

	apiKey, found, err := x.openRouterKey(ctx, thread.UserID)
	if err != nil {
		logger.Error("load openrouter key", "err", err)
		return
	}
	if !found {
		logger.Warn("skip resume: openrouter not connected")
		return
	}

	user, err := x.db.GetUser(ctx, thread.UserID)
	if err != nil {
		logger.Error("get user", "err", err)
		return
	}

	messages, err := x.db.GetMessagesByThread(ctx, thread.UserID, thread.ID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	messages = append(messages, db.Message{
		ThreadID: thread.ID,
		Type:     db.UserMessage,
		Content:  resumeMessage,
	})

	logger.Info("resuming agent stream")
	x.streamAgentMessages(ctx, apiKey, user.SelectedModel, thread.UserID, thread.ID, messages, logger)
}
