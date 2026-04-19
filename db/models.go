package db

import (
	"time"
)

// Thread is the public domain model for a conversation thread
type Thread struct {
	ID          string
	Title       string
	ArchivedAt  *time.Time // nil if not archived
	IsStreaming bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message is the public domain model for a chat message
type Message struct {
	ID                 string
	ThreadID           string
	Type               string // 'user_message', 'assistant_message', 'system_message', 'assistant_reasoning', 'assistant_tool_call', 'tool_output'
	Seq                int
	Content            string
	ToolCallID         string
	ToolName           string
	ProviderResponseID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
