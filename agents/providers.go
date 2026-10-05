package agents

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Strict      bool
}

//sumtype:decl
type StreamEvent interface {
	isStreamEvent()
}

type (
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

func (AssistantDelta) isStreamEvent() {}
func (ReasoningDelta) isStreamEvent() {}
func (Toolcall) isStreamEvent()       {}
func (ToolOutput) isStreamEvent()     {}
func (UsageDelta) isStreamEvent()     {}
