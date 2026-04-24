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

var _ = styler.Style(`
	.message-thread {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		min-height: 0;
		background: var(--bg-primary);
	}
	.thread-shell {
		display: flex;
		flex-direction: column;
		flex: 1;
		position: relative;
		min-height: 0;
	}
	.jump-to-bottom-btn {
		position: absolute;
		left: 50%;
		bottom: var(--size-4);
		transform: translateX(-50%);
		z-index: 10;
		inline-size: 44px;
		block-size: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		border: 1px solid color-mix(in oklch, var(--bg-button) 30%, var(--border-color));
		border-radius: var(--radius-round);
		background: color-mix(in oklch, var(--bg-primary) 88%, var(--bg-button));
		color: var(--text-primary);
		box-shadow: var(--shadow-3);
		cursor: pointer;
		transition:
			transform 0.15s var(--ease-3),
			filter 0.15s var(--ease-3),
			opacity 0.15s var(--ease-3);

		&:hover {
			filter: brightness(1.08);
			transform: translateX(-50%) translateY(-1px);
		}

		&:focus-visible {
			outline: none;
			box-shadow:
				var(--shadow-3),
				0 0 0 3px color-mix(in oklch, var(--bg-button) 28%, transparent);
		}

		svg {
			inline-size: 22px;
			block-size: 22px;
		}
	}
`)

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

	return Main(Class("message-thread"), contents)
}

func PatchThreadMessageList(
	sse *datastar.ServerSentEventGenerator, messages []MessageGroup, scrollSidebar bool) {
	ssePatch(sse, MessageThread{Messages: messages, DoScrollSidebar: scrollSidebar}.MessageList())
}

var _ = styler.Style(`
	.main-body {
		flex: 1;
		min-height: 0;
		min-width: 0;
		display: flex;
		flex-direction: column;
		align-items: stretch;
		padding: var(--size-5) var(--size-6);
		gap: var(--size-3);
		overflow-y: auto;
		overflow-x: hidden;
		overflow-anchor: auto;
		overscroll-behavior: contain;
		width: 100%;
		max-width: 768px;
		margin: 0 auto;
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
	}
`)

var _ = styler.Style(`
	.empty-thread {
		justify-content: flex-end;
		padding-top: var(--size-8);
	}
`)

var _ = styler.Style(`
	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: var(--size-3);
		text-align: center;
		color: var(--text-secondary);
	}

	.empty-state-title {
		font-size: clamp(2rem, 3vw, 2.75rem);
		line-height: 1.1;
		color: var(--text-primary);
		letter-spacing: -0.04em;
	}

	.empty-state-subtitle {
		max-width: 42ch;
		font-size: var(--font-size-1);
		line-height: var(--font-lineheight-3);
	}
`)

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

var _ = styler.Style(`
	.message {
		display: block;
		font-size: var(--font-size-1);
		line-height: var(--font-lineheight-4);
		padding: var(--size-3) var(--size-4);
		border-radius: var(--radius-3);
		word-break: break-word;
		min-width: 0;
		width: fit-content;
		max-width: min(100%, 72ch);
	}
	.user-message {
		background: color-mix(in oklch, var(--bg-button) 18%, transparent);
		border: 1px solid color-mix(in oklch, var(--bg-button) 30%, transparent);
		color: var(--text-primary);
		white-space: normal;
		margin-left: auto;
		margin-right: 0;
	}
`)

func (x UserMessage) Node() Node {
	return Div(
		ID("msg-"+x.ID),
		Class("message user-message"),
		MarkdownContent(x.MdContent),
	)
}

var _ = styler.Style(`
	.markdown-content {
		pre {
			background: var(--gray-11);
			border-radius: var(--radius-2);
			padding: var(--size-3);
			max-width: 100%;
			overflow-x: auto;
			margin: var(--size-2) 0;
		}

		code {
			font-family: 'Fira Code', var(--font-mono);
			font-size: var(--font-size-0);
		}

		:not(pre) > code {
			background: var(--gray-10);
			padding: var(--size-1) var(--size-2);
			border-radius: var(--radius-1);
		}

		p {
			margin: var(--size-2) 0;
		}
		p:first-child {
			margin-top: 0;
		}
		p:last-child {
			margin-bottom: 0;
		}

		ul,
		ol {
			padding-left: var(--size-5);
			margin: var(--size-2) 0;
		}

		li {
			margin: var(--size-1) 0;
		}

		h1,
		h2,
		h3,
		h4 {
			margin: var(--size-3) 0 var(--size-2);
			color: var(--text-primary);
		}

		h1 {
			font-size: var(--font-size-4);
		}
		h2 {
			font-size: var(--font-size-3);
		}
		h3 {
			font-size: var(--font-size-2);
		}

		blockquote {
			border-left: 3px solid var(--gray-8);
			padding-left: var(--size-3);
			color: var(--text-secondary);
			margin: var(--size-2) 0;
		}

		a {
			color: var(--bg-button);
			text-decoration: underline;
		}

		table {
			border-collapse: collapse;
			margin: var(--size-2) 0;
			width: 100%;
		}

		th,
		td {
			border: 1px solid var(--border-color);
			padding: var(--size-1) var(--size-2);
			text-align: left;
		}

		th {
			background: var(--gray-11);
		}

		hr {
			border: none;
			border-top: 1px solid var(--border-color);
			margin: var(--size-3) 0;
		}

		.chroma {
			background: var(--gray-11);
			color: var(--gray-3);
		}

		.chroma .k,
		.chroma .kd,
		.chroma .kn,
		.chroma .kc,
		.chroma .kr {
			color: var(--bg-button);
			font-weight: var(--font-weight-6);
		}

		.chroma .s,
		.chroma .sa,
		.chroma .sb,
		.chroma .sc,
		.chroma .sd,
		.chroma .s2,
		.chroma .se,
		.chroma .sh,
		.chroma .si,
		.chroma .sx,
		.chroma .sr,
		.chroma .s1,
		.chroma .ss {
			color: #7ee787;
		}

		.chroma .nf,
		.chroma .fm {
			color: var(--gray-1);
		}

		.chroma .c,
		.chroma .ch,
		.chroma .cm,
		.chroma .c1,
		.chroma .cs,
		.chroma .cp,
		.chroma .cpf {
			color: var(--gray-7);
			font-style: italic;
		}

		.chroma .nc,
		.chroma .kt,
		.chroma .nn {
			color: #ff7b72;
		}

		.chroma .mi,
		.chroma .mf,
		.chroma .mh,
		.chroma .mo,
		.chroma .mb,
		.chroma .il,
		.chroma .m {
			color: #79c0ff;
		}

		.chroma .nt,
		.chroma .na {
			color: #d2a8ff;
		}

		.chroma .nb,
		.chroma .bp {
			color: var(--gray-5);
		}

		.chroma .nv,
		.chroma .vc,
		.chroma .vg,
		.chroma .vi,
		.chroma .vm {
			color: var(--gray-4);
		}

		.chroma .o,
		.chroma .ow,
		.chroma .p {
			color: var(--gray-5);
		}
	}
`)

func MarkdownContent(mdContent string) Node {
	return Div(
		Class("markdown-content"),
		Raw(mdContent),
	)
}

var _ = styler.Style(`
	.assistant-message {
		color: var(--gray-3);
		white-space: normal;
		margin-left: 0;
		margin-right: auto;
		display: flex;
		flex-direction: column;
		gap: var(--size-3);

		&.large-height {
			min-height: calc(100vh - 13rem);
		}
	}
`)

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

var _ = styler.Style(`
	.assistant-part {
		min-width: 0;
	}

	.assistant-text-part {
		display: block;
	}
`)

func (x AssistantTextPart) Node() Node {
	return Div(
		Class("assistant-part assistant-text-part"),
		MarkdownContent(x.MdContent),
	)
}

var _ = styler.Style(`
	.reasoning-part,
	.tool-call-part {
		width: 100%;
		border: 1px solid var(--border-color);
		border-radius: var(--radius-2);
		background: var(--gray-11);
		overflow: hidden;

		summary {
			cursor: pointer;
			list-style: none;
			padding: var(--size-2) var(--size-3);
			font-size: var(--font-size-0);
			font-weight: var(--font-weight-6);
			color: var(--text-secondary);
		}

		summary::-webkit-details-marker {
			display: none;
		}
	}
`)

var _ = styler.Style(`
	.tool-call-part-summary {
		border-bottom: 1px solid var(--border-color);
	}

	.tool-call-part-body {
		margin: 0;
		border-radius: 0;
	}
`)

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

var _ = styler.Style(`
	.reasoning-part-body {
		padding: 0 var(--size-3) var(--size-3);
		font-size: var(--font-size-0);
		line-height: var(--font-lineheight-3);
		color: var(--text-secondary);
	}
`)

func (x ReasoningPart) Node() Node {
	return Details(
		Class("assistant-part reasoning-part"),
		Summary(
			Class("reasoning-part-summary"),
			Text("Reasoning"),
		),
		Div(
			Class("reasoning-part-body"),
			MarkdownContent(x.Text),
		),
	)
}
