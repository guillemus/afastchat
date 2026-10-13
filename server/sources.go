package server

import (
	"context"
	"encoding/json"
	"fmt"

	"afastchat.com/agents"
	"afastchat.com/db"
)

type turnSources struct {
	messageID string
	refs      []db.SourceRef
}

func (x *turnSources) add(
	ctx context.Context, pool *db.Pool, userID, threadID string, messages []db.Message, source agents.Source,
) ([]db.Message, error) {
	for _, ref := range x.refs {
		if ref.URL == source.URL {
			return messages, nil
		}
	}
	x.refs = append(x.refs, db.SourceRef{URL: source.URL, Title: source.Title})

	content, err := json.Marshal(x.refs)
	if err != nil {
		return messages, fmt.Errorf("marshal sources: %w", err)
	}

	if x.messageID != "" {
		return messages, pool.UpdateMessageContent(ctx, x.messageID, string(content))
	}

	created, err := pool.CreateMessageWithMeta(ctx, userID, db.Message{
		ThreadID: threadID,
		Type:     db.AssistantSources,
		Content:  string(content),
	})
	if err != nil {
		return messages, err
	}
	x.messageID = created.ID
	return append(messages, created), nil
}
