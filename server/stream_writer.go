package server

import (
	"context"
	"iter"
	"log/slog"

	"afastchat.com/agents"
	"afastchat.com/db"
)

type streamWriter struct {
	x        *Server
	ctx      context.Context
	logger   *slog.Logger
	userID   string
	threadID string
	messages []db.Message
	sources  turnSources
	dirty    bool
}

func (w *streamWriter) write(evt agents.StreamEvent) error {
	switch evt := evt.(type) {
	case agents.AssistantDelta:
		return w.appendDelta(db.AssistantMessage, evt.Delta)
	case agents.ReasoningDelta:
		return w.appendDelta(db.AssistantReasoning, evt.Delta)
	case agents.Toolcall:
		return w.create(db.Message{
			ThreadID:   w.threadID,
			Type:       db.AssistantToolCall,
			Content:    evt.Args,
			ToolCallID: evt.CallID,
			ToolName:   evt.Name,
		})
	case agents.ToolOutput:
		return w.create(db.Message{
			ThreadID:   w.threadID,
			Type:       db.ToolOutput,
			Content:    evt.Output,
			ToolCallID: evt.CallID,
			ToolName:   evt.Name,
		})
	case agents.Source:
		if err := w.flush(); err != nil {
			return err
		}
		messages, err := w.sources.add(w.ctx, w.x.db, w.userID, w.threadID, w.messages, evt)
		w.messages = messages
		if err != nil {
			return err
		}
		w.x.broker.PublishThreadUpdated(w.threadID)
		return nil
	case agents.UsageDelta:
		w.logger.Debug("received usage",
			"provider", evt.Usage.Provider,
			"model", evt.Usage.Model,
			"input_tokens", evt.Usage.InputTokens,
			"cached_input_tokens", evt.Usage.CachedInputTokens,
			"output_tokens", evt.Usage.OutputTokens,
			"reasoning_tokens", evt.Usage.ReasoningTokens,
		)
		_, err := w.x.db.CreateUsage(w.ctx, db.Usage{
			UserID:            w.userID,
			ThreadID:          w.threadID,
			MessageID:         lastAssistantMessageID(w.messages),
			Provider:          evt.Usage.Provider,
			Model:             evt.Usage.Model,
			InputTokens:       evt.Usage.InputTokens,
			CachedInputTokens: evt.Usage.CachedInputTokens,
			OutputTokens:      evt.Usage.OutputTokens,
			ReasoningTokens:   evt.Usage.ReasoningTokens,
		})
		return err
	}
	return nil
}

func (w *streamWriter) appendDelta(typ db.MessageType, delta string) error {
	if n := len(w.messages); n > 0 && w.messages[n-1].Type == typ {
		w.messages[n-1].Content += delta
		w.dirty = true
		return nil
	}

	return w.create(db.Message{ThreadID: w.threadID, Type: typ, Content: delta})
}

func (w *streamWriter) create(message db.Message) error {
	if err := w.flush(); err != nil {
		return err
	}

	created, err := w.x.db.CreateMessageWithMeta(w.ctx, w.userID, message)
	if err != nil {
		return err
	}
	w.messages = append(w.messages, created)
	w.x.broker.PublishThreadUpdated(w.threadID)
	return nil
}

func (w *streamWriter) flush() error {
	if !w.dirty {
		return nil
	}

	last := w.messages[len(w.messages)-1]
	if err := w.x.db.UpdateMessageContent(w.ctx, last.ID, last.Content); err != nil {
		return err
	}
	w.dirty = false
	w.x.broker.PublishThreadUpdated(w.threadID)
	return nil
}

type streamItem struct {
	evt agents.StreamEvent
	err error
}

func pullStream(ctx context.Context, stream iter.Seq2[agents.StreamEvent, error]) <-chan streamItem {
	out := make(chan streamItem)
	GoRecover(func() {
		defer close(out)
		for evt, err := range stream {
			select {
			case out <- streamItem{evt: evt, err: err}:
			case <-ctx.Done():
				return
			}
		}
	})
	return out
}
