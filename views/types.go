package views

type Thread struct {
	ID          string
	IsStreaming bool
	Title       string
}

type IndexData struct {
	Threads   []Thread
	UserEmail string
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []MessageGroup
	Streaming bool
	UserEmail string
}

type SettingsData struct {
	Threads   []Thread
	UserEmail string
}

//sumtype:decl
type MessageGroup interface {
	isMessageGroup()
}

type UserMessageProps struct {
	ID        string
	MdContent string
}

type AssistantMessagePartsProps struct {
	ID          string
	LargeHeight bool
	Parts       []AssistantPart
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
	Text      string
	MdContent string
}

func (AssistantTextPartProps) isAssistantPart() {}
func (ToolCallPartProps) isAssistantPart()      {}
func (ReasoningPartProps) isAssistantPart()     {}

type MessageThreadProps struct {
	Messages        []MessageGroup
	ThreadID        string
	DoScrollSidebar bool
	IsStreaming     bool
}
