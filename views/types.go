package views

import "time"

type Thread struct {
	ID          string
	IsStreaming bool
	IsPinned    bool
	Title       string
	CreatedAt   time.Time
}

type Viewer struct {
	Name  string
	Email string
}

type IndexData struct {
	Threads []Thread
	Viewer  Viewer
	Models  ModelPicker
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []MessageGroup
	Streaming bool
	Viewer    Viewer
	Models    ModelPicker
}

type ModelOption struct {
	ID          string
	Name        string
	Provider    string
	Description string
	CostTier    int
}

type ModelProvider struct {
	ID    string
	Label string
}

type ModelPicker struct {
	Providers []ModelProvider
	Options   []ModelOption
	Selected  string
	Favorites []string
}

type SettingsData struct {
	Threads             []Thread
	Viewer              Viewer
	OpenRouterConnected bool
	OpenRouterFailed    bool
}

//sumtype:decl
type MessageGroup interface {
	isMessageGroup()
}

type UserMessageProps struct {
	ID        string
	MdContent string
	CreatedAt time.Time
}

type AssistantMessagePartsProps struct {
	ID          string
	LargeHeight bool
	Working     bool
	Streaming   bool
	StartedAt   time.Time
	WorkSeconds int
	Work        []AssistantPart
	AnswerHTML  string
	Sources     []Source
	Status      Status
}

type Source struct {
	Number int
	URL    string
	Title  string
	Domain string
}

type StatusIcon uint8

const (
	StatusNoIcon StatusIcon = iota
	StatusThinking
	StatusSearching
	StatusReading
)

type Status struct {
	Key    string
	Icon   StatusIcon
	Verb   string
	Target string
	Words  string
}

func (UserMessageProps) isMessageGroup()           {}
func (AssistantMessagePartsProps) isMessageGroup() {}

//sumtype:decl
type AssistantPart interface {
	isAssistantPart()
}

type AssistantTextPartProps struct {
	ID        string
	MdContent string
}

type ToolCallPartProps struct {
	ID   string
	Name string
	Args string
}

type ReasoningPartProps struct {
	ID        string
	MdContent string
	Seconds   int
}

type SourceReadPartProps struct {
	ID     string
	Title  string
	Domain string
	URL    string
}

func (AssistantTextPartProps) isAssistantPart() {}
func (ToolCallPartProps) isAssistantPart()      {}
func (ReasoningPartProps) isAssistantPart()     {}
func (SourceReadPartProps) isAssistantPart()    {}

type MessageThreadProps struct {
	Messages        []MessageGroup
	Thread          Thread
	DoScrollSidebar bool
	IsStreaming     bool
	Models          ModelPicker
}
