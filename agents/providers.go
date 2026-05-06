package agents

import (
	"context"
	"iter"
	"strings"
)

type Provider interface {
	Stream(context.Context, ProviderRequest) iter.Seq2[StreamEvent, error]
}

type ProviderRequest struct {
	Messages     []Message
	Tools        []Tool
	Instructions string
	State        ProviderState
}

type ProviderState struct {
	OpenAIResponseID string
}

type Message struct {
	Role       string
	Content    string
	ToolCallID string
	ToolName   string
}

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Strict      bool
}

type (
	StreamEvent interface {
		isStreamEvent()
	}
	AssistantDelta struct {
		Delta string
	}
	ReasoningDelta struct {
		Delta string
	}
	Toolcall struct {
		CallID string
		Name   string
		Args   string
	}
	ToolOutput struct {
		CallID string
		Name   string
		Output string
	}
	ProviderStateDelta struct {
		State ProviderState
	}
	Usage struct {
		Provider          string
		Model             string
		InputTokens       int64
		CachedInputTokens int64
		OutputTokens      int64
		ReasoningTokens   int64
	}
	UsageDelta struct {
		Usage Usage
	}
)

func (AssistantDelta) isStreamEvent()     {}
func (ReasoningDelta) isStreamEvent()     {}
func (Toolcall) isStreamEvent()           {}
func (ToolOutput) isStreamEvent()         {}
func (ProviderStateDelta) isStreamEvent() {}
func (UsageDelta) isStreamEvent()         {}

type ToolcallBuf struct {
	ItemID string
	CallID string
	Name   string
	Args   strings.Builder
}

type CalledTools struct {
	Tools map[string]*ToolcallBuf
	Order []string
}

func newCalledTools() *CalledTools {
	return &CalledTools{Tools: map[string]*ToolcallBuf{}}
}

func (x *CalledTools) functionCall(itemID string) *ToolcallBuf {
	call := x.Tools[itemID]
	if call != nil {
		return call
	}

	call = &ToolcallBuf{ItemID: itemID}
	x.Tools[itemID] = call
	x.Order = append(x.Order, itemID)

	return call
}
