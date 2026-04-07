package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// here all queries that wrap sqlc generated queries with a nicer API

func (c *Conn) CreateUserMessage(
	ctx context.Context,
	threadID uuid.UUID, msg string,
) (Message, error) {
	return c.Queries.CreateMessage(ctx, CreateMessageParams{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "user",
		Content:  msg,
	})
}

func (c *Conn) CreateAssistantMessage(
	ctx context.Context,
	threadID uuid.UUID, msg string,
) (uuid.UUID, error) {
	createdMsg, err := c.Queries.CreateMessage(ctx, CreateMessageParams{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "assistant",
		Content:  msg,
	})

	return createdMsg.ID, err
}

func (c *Conn) CreateSystemMessage(
	ctx context.Context,
	threadID uuid.UUID, msg string,
) (uuid.UUID, error) {
	createdMsg, err := c.Queries.CreateMessage(ctx, CreateMessageParams{
		ID:       uuid.New(),
		ThreadID: threadID,
		Role:     "system",
		Content:  msg,
	})

	return createdMsg.ID, err
}

// InitThread creates a thread with messages prefilled
func (c *Conn) InitThread(
	ctx context.Context, title, sysMsg, userMsg string) (*uuid.UUID, error) {
	thread, err := c.Queries.CreateThread(ctx, uuid.New(), title)
	if err != nil {
		return nil, fmt.Errorf("create thread: %w", err)
	}

	_, err = c.CreateSystemMessage(ctx, thread.ID, sysMsg)
	if err != nil {
		return nil, fmt.Errorf("create system message: %w", err)
	}

	_, err = c.CreateUserMessage(ctx, thread.ID, userMsg)
	if err != nil {
		return nil, fmt.Errorf("create user message: %w", err)
	}

	return &thread.ID, nil
}
