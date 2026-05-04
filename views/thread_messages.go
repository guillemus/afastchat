package views

import (
	"log/slog"

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
		ID   string
		Text string
	}
	ReasoningPart struct {
		ID   string
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

var _ = styles.Style(`
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
		scrollbar-gutter: stable;
		overflow-anchor: auto;
		overscroll-behavior: contain;
		width: 100%;
		max-width: 960px;
		margin: 0 auto;
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
	}
	.empty-thread {
		justify-content: flex-end;
		padding-top: var(--size-8);
	}
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
		line-height: 1.1;
		color: var(--text-primary);
		letter-spacing: -0.04em;
	}

	.empty-state-subtitle {
		max-width: 42ch;
		line-height: var(--font-lineheight-3);
	}
	.message {
		flex: 0 0 auto;
		line-height: var(--font-lineheight-4);
		padding: var(--size-3) var(--size-4);
		border-radius: var(--radius-3);
		word-break: break-word;
		min-width: 0;
		width: fit-content;
		max-width: min(100%, 72ch);
	}
	.user-message {
		max-width: min(100%, 56ch);
		padding: var(--size-2) var(--size-3);
		border-radius: var(--radius-2);
		background: color-mix(in oklch, var(--gray-10) 72%, transparent);
		border: 1px solid var(--border-color);
		color: var(--text-primary);
		white-space: normal;
		margin-left: auto;

		.markdown-content p {
			margin: 0;
		}
	}
	.assistant-message {
		width: 100%;
		max-width: 100%;
		color: var(--gray-3);
		white-space: normal;
		margin-right: auto;
		display: flex;
		flex-direction: column;
		gap: var(--size-3);

		&.large-height {
			min-height: calc(100vh - 13rem);
		}
	}
	.assistant-part {
		min-width: 0;
		flex: 0 0 auto;
	}

	.assistant-text-part {
		width: 100%;
	}
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
			font-weight: var(--font-weight-6);
			color: var(--text-secondary);
		}

		summary::-webkit-details-marker {
			display: none;
		}
	}
	.tool-call-part-summary {
		border-bottom: 1px solid var(--border-color);
	}

	.tool-call-part-body {
		border-radius: 0;
	}
	.reasoning-part-body {
		padding: 0 var(--size-3) var(--size-3);
		line-height: var(--font-lineheight-3);
		color: var(--text-secondary);
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

	return Main(
		Class("message-thread"),
		contents,
		If(x.DoScrollSidebar && len(x.Messages) > 0, MessageScrollRestoreScript()),
	)
}

func MessageScrollRestoreScript() Node {
	// Keep this after all message-thread layout-affecting nodes, especially chat input.
	// ScrollTop is clamped to the current max scroll; restoring before input is parsed
	// uses a too-tall message list, then input shrinks it and leaves scroll above bottom.
	return Script(Raw(`{
		let el = document.querySelector('#messages');
		el.scrollTop = el.scrollHeight;
	}`))
}

func PatchThreadMessageList(
	sse *datastar.ServerSentEventGenerator, messages []MessageGroup, scrollMessages bool) {
	ssePatch(sse, MessageThread{Messages: messages}.MessageList())
	if scrollMessages {
		err := sse.ExecuteScript(`
			let el = document.querySelector('#messages');
			el.scrollTop = el.scrollHeight;
		`)
		if err != nil {
			slog.Error("sse.ExecuteScript error", "err", err)
		}
	}
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
		Class("main-body"),
		Data("on:scroll__debounce.10ms", `
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`),

		Map(x.Messages, func(m MessageGroup) Node {
			return m.Node()
		}),
	)
}

func (x UserMessage) Node() Node {
	return Div(
		ID("msg-"+x.ID),
		Class("message user-message"),
		MarkdownContent(x.MdContent),
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
		ID("part-"+x.ID),
		Class("assistant-part assistant-text-part"),
		MarkdownContent(x.MdContent),
	)
}

func (x ToolCallPart) Node() Node {
	return Details(
		ID("part-"+x.ID),
		Class("assistant-part tool-call-part"),
		Data("preserve-attr", "open"),
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
		ID("part-"+x.ID),
		Class("assistant-part reasoning-part"),
		Data("preserve-attr", "open"),
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
