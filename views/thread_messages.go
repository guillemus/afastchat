package views

import (
	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type (
	MessageGroup interface {
		isMessageGroup()
		Node() Node
	}
	UserMessage struct {
		ID        string
		MdContent string
	}
	AssistantMessageParts struct {
		ID          string
		LargeHeight bool
		Parts       []AssistantPart
	}
)

func (UserMessage) isMessageGroup()           {}
func (AssistantMessageParts) isMessageGroup() {}

type (
	AssistantPart interface {
		isAssistantPart()
		Node() Node
	}
	AssistantTextPart struct {
		ID        string
		MdContent string
	}
	ToolCallPart struct {
		Text string
	}
	ReasoningPart struct {
		Text string
	}
)

func (AssistantTextPart) isAssistantPart() {}
func (ToolCallPart) isAssistantPart()      {}
func (ReasoningPart) isAssistantPart()     {}

type MessageThread struct {
	Messages        []MessageGroup
	ThreadID        string
	DoScrollSidebar bool
	IsStreaming     bool
}

func NewEmptyMessageThread() Node {
	return MessageThread{
		Messages:        []MessageGroup{},
		ThreadID:        "",
		DoScrollSidebar: false,
		IsStreaming:     false,
	}.Node()
}

func (x MessageThread) Node() Node {
	var contents Node
	if len(x.Messages) == 0 {
		input := newThreadChatInput()
		if x.ThreadID != "" {
			input = existingThreadChatInput(x.ThreadID, x.IsStreaming)
		}

		contents = Group{
			x.MessageList(),
			input,
		}
	} else {
		contents = Group{
			Div(
				Class("thread-shell"),
				Data("signals:show-jump-to-bottom", "false"),
				Data("on:resize__window__debounce.50ms", `
					if (!$messages) return;
					const overflow = $messages.scrollHeight - $messages.clientHeight;
					const distanceFromBottom = overflow - $messages.scrollTop;
					$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
				`),
				x.MessageList(),
				Button(
					Type("button"),
					Style("display: none"),
					Class("jump-to-bottom-btn"),
					Data("show", "$showJumpToBottom"),
					Data("on:click", `
						$messages.scrollTo({ top: $messages.scrollHeight, behavior: 'smooth' });
						$showJumpToBottom = false;
					`),
					iconArrowDown(),
				),
			),
			existingThreadChatInput(x.ThreadID, x.IsStreaming),
		}
	}

	return Main(Class("main"), contents)
}

func PatchThreadMessageList(
	sse *datastar.ServerSentEventGenerator, messages []MessageGroup, scrollSidebar bool) {
	ssePatch(sse, MessageThread{Messages: messages, DoScrollSidebar: scrollSidebar}.MessageList())
}

func (x MessageThread) MessageList() Node {
	if len(x.Messages) == 0 {
		return Div(
			ID("messages"),
			Class("main-body empty-thread"),
			Div(
				Class("empty-state"),
				H1(Class("empty-state-title"), Text("How can I help?")),
				P(Class("empty-state-subtitle"),
					Text("Start a new conversation with afastchat."),
				),
			),
		)
	}

	return Div(
		ID("messages"),
		Data("ref:messages", ""),
		Classes{
			"main-body": true,
			"invisible": x.DoScrollSidebar,
		},
		Data("on:scroll__debounce.10ms", `
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`),
		If(x.DoScrollSidebar, Data("init", `
			el.scrollTop = el.scrollHeight;
			el.classList.remove('invisible');
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`)),
		Map(x.Messages, func(m MessageGroup) Node {
			return m.Node()
		}),
	)
}

func (x UserMessage) Node() Node {
	return Div(
		ID("msg-"+x.ID),
		Class("message user-message"),
		Raw(x.MdContent),
	)
}

func (x AssistantMessageParts) Node() Node {
	return x.assistantMessage()
}

func (x AssistantMessageParts) assistantMessage() Node {
	return Div(
		ID("msg-"+x.ID),
		Classes{
			"message assistant-message": true,
			"large-height":              x.LargeHeight,
		},
		Map(x.Parts, func(part AssistantPart) Node {
			return part.Node()
		}),
	)
}

func (x AssistantTextPart) Node() Node {
	return Div(
		Class("assistant-part assistant-text-part"),
		Raw(x.MdContent),
	)
}

func (x ToolCallPart) Node() Node {
	return Details(
		Class("assistant-part tool-call-part"),
		Summary(
			Class("tool-call-part-summary"),
			Text("Tool call"),
		),
		Pre(
			Class("tool-call-part-body"),
			Text(x.Text),
		),
	)
}

func (x ReasoningPart) Node() Node {
	return Details(
		Class("assistant-part reasoning-part"),
		Summary(
			Class("reasoning-part-summary"),
			Text("Reasoning"),
		),
		Div(
			Class("reasoning-part-body"),
			Raw(x.Text),
		),
	)
}
