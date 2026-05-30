package views

import (
	"encoding/json"
	"strconv"
	"strings"

	"afastchat.com/agents"
	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type (
	MessageGroup interface {
		isMessageGroup()
	}
	UserMessageProps struct {
		ID        string
		MdContent string
	}
	AssistantMessagePartsProps struct {
		ID          string
		LargeHeight bool
		Parts       []AssistantPart
	}
)

func (UserMessageProps) isMessageGroup()           {}
func (AssistantMessagePartsProps) isMessageGroup() {}

type (
	AssistantPart interface {
		isAssistantPart()
	}
	AssistantTextPartProps struct {
		ID        string
		MdContent string
	}
	ToolCallPartProps struct {
		ID   string
		Name string
		Args string
	}
	ReasoningPartProps struct {
		ID        string
		Text      string
		MdContent string
	}
)

func (AssistantTextPartProps) isAssistantPart() {}
func (ToolCallPartProps) isAssistantPart()      {}
func (ReasoningPartProps) isAssistantPart()     {}

type MessageThreadProps struct {
	Messages        []MessageGroup
	ThreadID        string
	DoScrollSidebar bool
	IsStreaming     bool
}

func NewEmptyMessageThread() Node {
	return MessageThread(MessageThreadProps{
		Messages:        []MessageGroup{},
		ThreadID:        "",
		DoScrollSidebar: false,
		IsStreaming:     false,
	})
}

func MessageThread(x MessageThreadProps) Node {
	var contents Node
	if len(x.Messages) == 0 {
		input := newThreadChatInput()
		if x.ThreadID != "" {
			input = existingThreadChatInput(x.ThreadID, x.IsStreaming)
		}

		contents = Group{
			MessageList(x),
			input,
		}
	} else {
		contents = Group{
			Div(
				Class("thread-shell relative flex min-h-0 flex-1 flex-col overflow-hidden"),
				Data("signals:show-jump-to-bottom", "false"),
				Data("on:resize__window__debounce.50ms", `
					if (!$messages) return;
					const overflow = $messages.scrollHeight - $messages.clientHeight;
					const distanceFromBottom = overflow - $messages.scrollTop;
					$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
				`),
				MessageList(x),
				Button(
					Type("button"),
					Style("display: none"),
					Class("absolute left-1/2 bottom-4 z-10 flex size-11 -translate-x-1/2 cursor-pointer items-center justify-center rounded-full border border-[color-mix(in_oklch,#f97316_30%,#262626)] bg-[color-mix(in_oklch,#0a0a0a_88%,#f97316)] text-neutral-100 shadow-[0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)] transition hover:-translate-y-px hover:brightness-110 focus-visible:outline-none focus-visible:shadow-[0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1),0_0_0_3px_color-mix(in_oklch,#f97316_28%,transparent)]"),
					Data("show", "$showJumpToBottom"),
					Data("on:click", `
						$messages.scrollTo({ top: $messages.scrollHeight, behavior: 'smooth' });
						$showJumpToBottom = false;
					`),
					iconArrowDown("size-[22px]"),
				),
			),
			existingThreadChatInput(x.ThreadID, x.IsStreaming),
		}
	}

	return Main(
		Class("message-thread flex h-full min-h-0 min-w-0 flex-col overflow-hidden bg-neutral-950"),
		contents,
		If(x.DoScrollSidebar && len(x.Messages) > 0, MessageScrollRestoreScript),
	)
}

// Keep this after all message-thread layout-affecting nodes, especially chat input.
// ScrollTop is clamped to the current max scroll; restoring before input is parsed
// uses a too-tall message list, then input shrinks it and leaves scroll above bottom.
var MessageScrollRestoreScript = Static(Script(Raw(`{
	let el = document.querySelector('#messages');
	el.scrollTop = el.scrollHeight;
}`)))

var EmptyMessageState = Static(Div(
	Class("flex flex-1 flex-col items-center justify-center gap-3 text-center text-neutral-400"),
	H1(Class("leading-tight tracking-[-0.04em] text-neutral-100"), Text("How can I help?")),
	P(Class("max-w-[42ch] leading-7"),
		Text("Start a new conversation with afastchat."),
	),
))

func PatchThreadMessageList(sse *datastar.ServerSentEventGenerator, messages []MessageGroup, scrollMessages bool) {
	ssePatch(sse, MessageList(MessageThreadProps{Messages: messages}))
	if scrollMessages {
		sseExecJS(sse, `
			let el = document.querySelector('#messages');
			let userMessages = el?.querySelectorAll('.user-message');
			let latestUserMessage = userMessages?.[userMessages.length - 1];
			if (latestUserMessage) {
				latestUserMessage.scrollIntoView({ block: 'start' });
			} else if (el) {
				el.scrollTop = el.scrollHeight;
			}
		`)
	}
}

func MessageList(x MessageThreadProps) Node {
	if len(x.Messages) == 0 {
		return Div(
			ID("messages"),
			Class("main-body flex min-h-0 min-w-0 flex-1 flex-col items-stretch gap-3 overflow-y-auto overflow-x-hidden overscroll-contain px-5 py-5 pt-8 justify-end mx-auto w-full max-w-[860px] [overflow-anchor:auto] [scrollbar-color:theme(colors.neutral.700)_transparent] [scrollbar-gutter:stable] [scrollbar-width:thin]"),
			EmptyMessageState,
		)
	}

	return Div(
		ID("messages"),
		Data("ref:messages", ""),
		Class("main-body mx-auto flex min-h-0 min-w-0 w-full max-w-[860px] flex-1 flex-col items-stretch gap-3 overflow-y-auto overflow-x-hidden overscroll-contain px-5 py-5 [overflow-anchor:auto] [scrollbar-color:theme(colors.neutral.700)_transparent] [scrollbar-gutter:stable] [scrollbar-width:thin]"),
		Data("on:scroll__debounce.10ms", `
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`),

		Map(x.Messages, func(m MessageGroup) Node {
			return MessageGroupNode(m)
		}),
	)
}

func MessageGroupNode(x MessageGroup) Node {
	switch x := x.(type) {
	case UserMessageProps:
		return UserMessage(x)
	case AssistantMessagePartsProps:
		return AssistantMessageParts(x)
	}
	panic("unknown message group")
}

func UserMessage(x UserMessageProps) Node {
	return Div(
		ID("msg-"+x.ID),
		Class("message user-message ml-auto w-fit max-w-[min(100%,56ch)] min-w-0 shrink-0 break-words whitespace-normal rounded-lg border border-neutral-800 bg-[color-mix(in_oklch,#262626_72%,transparent)] px-3 py-2 leading-7 text-neutral-100"),
		MarkdownContent(x.MdContent),
	)
}

func AssistantMessageParts(x AssistantMessagePartsProps) Node {
	return Div(
		ID("msg-"+x.ID),
		Classes{
			"message assistant-message mr-auto flex w-full max-w-full shrink-0 flex-col gap-3 whitespace-normal break-words rounded-xl px-4 py-3 leading-7 text-neutral-200": true,
			"min-h-[calc(100dvh-13rem)]": x.LargeHeight,
		},
		Map(x.Parts, func(part AssistantPart) Node {
			return AssistantPartNode(part)
		}),
	)
}

func AssistantPartNode(x AssistantPart) Node {
	switch x := x.(type) {
	case AssistantTextPartProps:
		return AssistantTextPart(x)
	case ToolCallPartProps:
		return ToolCallPart(x)
	case ReasoningPartProps:
		return ReasoningPart(x)
	}
	panic("unknown assistant part")
}

func AssistantTextPart(x AssistantTextPartProps) Node {
	return Div(
		ID("part-"+x.ID),
		Class("assistant-part assistant-text-part min-w-0 shrink-0 w-full"),
		MarkdownContent(x.MdContent),
	)
}

func ToolCallPart(x ToolCallPartProps) Node {
	name, arg := toolCallDisplay(x.Name, x.Args)
	return Div(
		ID("part-"+x.ID),
		Class("assistant-part tool-call-part flex w-full min-w-0 shrink-0 items-baseline gap-2 overflow-hidden rounded-lg border border-neutral-800 bg-neutral-900 px-3 py-2"),
		Span(Class("whitespace-nowrap font-semibold text-neutral-100"), Text(name)),
		Span(Class("min-w-0 overflow-hidden text-ellipsis whitespace-nowrap font-mono text-sm text-neutral-400"), Text(arg)),
	)
}

func ReasoningPart(x ReasoningPartProps) Node {
	return Details(
		ID("part-"+x.ID),
		Class("assistant-part reasoning-part w-full min-w-0 shrink-0 overflow-hidden rounded-lg border border-neutral-800 bg-neutral-900"),
		Data("preserve-attr", "open"),
		Summary(
			Class("flex min-w-0 cursor-pointer list-none items-baseline gap-2 px-3 py-2 text-neutral-400"),
			Span(Class("whitespace-nowrap font-semibold text-neutral-100"), Text("Reasoning")),
			Span(Class("min-w-0 overflow-hidden text-ellipsis whitespace-nowrap text-neutral-400"), Text(reasoningPreview(x.Text))),
		),
		Div(
			Class("px-3 pb-3 leading-7 text-neutral-400"),
			MarkdownContent(x.MdContent),
		),
	)
}

func reasoningPreview(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func toolCallDisplay(name, args string) (string, string) {
	switch name {
	case "web_search":
		var a agents.WebSearchArgs
		_ = json.Unmarshal([]byte(args), &a)
		return "Searching web", a.Query
	case "web_search_recent":
		var a agents.WebSearchRecentArgs
		_ = json.Unmarshal([]byte(args), &a)
		if a.DaysBack == 0 {
			return "Searching recent web", a.Query
		}
		return "Searching recent web", a.Query + " · " + strconv.Itoa(a.DaysBack) + " days"
	case "web_read":
		var a agents.WebReadArgs
		_ = json.Unmarshal([]byte(args), &a)
		if len(a.URLs) == 1 {
			return "Reading web page", a.URLs[0]
		}
		if len(a.URLs) > 1 {
			return "Reading web pages", strconv.Itoa(len(a.URLs)) + " URLs"
		}
		return "Reading web page", ""
	default:
		return name, ""
	}
}
