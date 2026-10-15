package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"afastchat.com/db"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
)

const chatTurnWorkflowPrefix = "chat-turn:"

type ChatTurnInput struct {
	UserID         string
	ThreadID       string
	UserMessageSeq int
	ModelID        string
}

func chatTurnWorkflowID(threadID, userMessageID string) string {
	return chatTurnWorkflowPrefix + threadID + ":" + userMessageID
}

func (x *Server) startChatTurn(
	ctx context.Context, userID, threadID, modelID, content string, attachments []db.AttachmentInput,
) (err error) {
	if err := x.db.SetThreadStreaming(ctx, userID, threadID, true); err != nil {
		return fmt.Errorf("set streaming state: %w", err)
	}
	defer func() {
		if err != nil {
			x.finishChatTurn(userID, threadID)
		}
	}()

	message, err := x.db.CreateUserMessage(ctx, userID, threadID, content)
	if err != nil {
		return fmt.Errorf("create user message: %w", err)
	}
	if err := x.db.CreateMessageAttachments(ctx, userID, message.ID, attachments); err != nil {
		return fmt.Errorf("create message attachments: %w", err)
	}

	_, err = dbos.RunWorkflow(x.workflows, x.runChatTurn, ChatTurnInput{
		UserID:         userID,
		ThreadID:       threadID,
		UserMessageSeq: message.Seq,
		ModelID:        modelID,
	}, dbos.WithWorkflowID(chatTurnWorkflowID(threadID, message.ID)))
	if err != nil {
		return fmt.Errorf("start chat turn workflow: %w", err)
	}

	x.broker.PublishUserThread(userID, threadID)
	x.broker.PublishThread(threadID, threadEvent{UserMessageCreated: true})
	return nil
}

func (x *Server) runChatTurn(ctx dbos.Context, input ChatTurnInput) (struct{}, error) {
	return dbos.RunAsStep(ctx, func(ctx context.Context) (struct{}, error) {
		x.streamChatTurn(ctx, input)
		return struct{}{}, nil
	}, dbos.WithStepName("stream"))
}

func (x *Server) streamChatTurn(ctx context.Context, input ChatTurnInput) {
	logger := slog.With("threadId", input.ThreadID, "userId", input.UserID)
	defer x.finishChatTurn(input.UserID, input.ThreadID)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := x.cancelOnThreadStop(ctx, input.ThreadID, cancel); err != nil {
		logger.Error("subscribe to thread stops", "err", err)
	}

	if err := x.db.DeleteMessagesAfter(ctx, input.UserID, input.ThreadID, input.UserMessageSeq); err != nil {
		logger.Error("delete partial turn output", "err", err)
		return
	}

	apiKey, found, err := x.openRouterKey(ctx, input.UserID)
	if err != nil {
		logger.Error("load openrouter key", "err", err)
		return
	}
	if !found {
		logger.Warn("skip chat turn: openrouter not connected")
		return
	}

	if _, err := x.db.CreateAssistantMessage(ctx, input.UserID, input.ThreadID, ""); err != nil {
		logger.Error("create assistant message", "err", err)
		return
	}
	x.broker.PublishThreadUpdated(input.ThreadID)

	messages, err := x.getThreadMessages(ctx, input.UserID, input.ThreadID)
	if err != nil {
		logger.Error("get messages", "err", err)
		return
	}

	logger.Info("starting agent stream")
	x.streamAgentMessages(ctx, apiKey, input.ModelID, input.UserID, input.ThreadID, messages, logger)
}

func (x *Server) finishChatTurn(userID, threadID string) {
	if err := x.db.SetThreadStreaming(context.Background(), userID, threadID, false); err != nil {
		slog.Error("unset streaming state", "threadId", threadID, "err", err)
		return
	}
	x.broker.PublishThreadUpdated(threadID)
	x.broker.PublishUserThread(userID, threadID)
}

func (x *Server) pendingChatTurns(threadID string) ([]string, error) {
	workflows, err := dbos.ListWorkflows(x.workflows,
		dbos.WithFilterWorkflowIDPrefix(chatTurnWorkflowPrefix+threadID),
		dbos.WithFilterStatus(dbos.WorkflowStatusPending, dbos.WorkflowStatusEnqueued),
	)
	if err != nil {
		return nil, fmt.Errorf("list chat turn workflows: %w", err)
	}

	ids := make([]string, 0, len(workflows))
	for _, workflow := range workflows {
		ids = append(ids, workflow.ID)
	}
	return ids, nil
}

func (x *Server) cancelChatTurns(threadID string) {
	ids, err := x.pendingChatTurns(threadID + ":")
	if err != nil {
		slog.Error("cancel chat turns", "threadId", threadID, "err", err)
		return
	}
	for _, id := range ids {
		if err := dbos.CancelWorkflow(x.workflows, id); err != nil {
			slog.Error("cancel chat turn workflow", "workflowId", id, "err", err)
		}
	}
}

func (x *Server) releaseOrphanedStreams() {
	ctx := context.Background()
	threads, err := x.db.ListStreamingThreads(ctx)
	if err != nil {
		slog.Error("list streaming threads", "err", err)
		return
	}
	if len(threads) == 0 {
		return
	}

	pending, err := x.pendingChatTurns("")
	if err != nil {
		slog.Error("release orphaned streams", "err", err)
		return
	}
	active := make(map[string]bool, len(pending))
	for _, id := range pending {
		threadID, _, _ := strings.Cut(strings.TrimPrefix(id, chatTurnWorkflowPrefix), ":")
		active[threadID] = true
	}

	for _, thread := range threads {
		if active[thread.ID] {
			continue
		}
		slog.Warn("releasing orphaned streaming thread", "threadId", thread.ID)
		x.finishChatTurn(thread.UserID, thread.ID)
	}
}
