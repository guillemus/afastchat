package db

import (
	"time"
)

type Thread struct {
	ID          string
	UserID      string
	Title       string
	ArchivedAt  *time.Time // nil if not archived
	PinnedAt    *time.Time
	IsStreaming bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MessageType string

const (
	UserMessage        MessageType = "user_message"
	AssistantMessage   MessageType = "assistant_message"
	SystemMessage      MessageType = "system_message"
	AssistantReasoning MessageType = "assistant_reasoning"
	AssistantToolCall  MessageType = "assistant_tool_call"
	ToolOutput         MessageType = "tool_output"
	AssistantSources   MessageType = "assistant_sources"
)

type SourceRef struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Message struct {
	ID                 string
	ThreadID           string
	Type               MessageType
	Seq                int
	Content            string
	ToolCallID         string
	ToolName           string
	ProviderResponseID string
	Attachments        []Attachment
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type User struct {
	ID            string
	Email         string
	DisplayName   string
	SelectedModel string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Usage struct {
	ID                string
	UserID            string
	ThreadID          string
	MessageID         string
	Provider          string
	Model             string
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
	ReasoningTokens   int64
	CreatedAt         time.Time
}
