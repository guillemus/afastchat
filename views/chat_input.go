package views

import (
	"fmt"

	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.chat-input {
		flex-shrink: 0;
		padding: var(--size-2) var(--size-6) var(--size-5);
		width: 100%;
		max-width: 768px;
		margin: 0 auto;
		
		@keyframes chat-input-pulse-border {
			0%,
			100% {
				box-shadow: 0 0 0 2px color-mix(in oklch, var(--bg-button) 10%, transparent);
			}
			50% {
				box-shadow: 0 0 0 4px color-mix(in oklch, var(--bg-button) 30%, transparent);
			}
		}

		.chat-input-box {
			display: flex;
			align-items: center;
			background: var(--bg-input);
			border: 1px solid var(--border-color);
			border-radius: var(--radius-3);
			padding: var(--size-3);

			&.is-sending {
				border-color: color-mix(in oklch, var(--bg-button) 60%, transparent);
				box-shadow: 0 0 0 2px color-mix(in oklch, var(--bg-button) 20%, transparent);
				animation: chat-input-pulse-border 1.5s ease-in-out infinite;
			}

			textarea {
				width: 100%;
				background: none;
				border: none;
				color: var(--text-primary);
				font-family: inherit;
				font-size: inherit;
				resize: none;
				outline: none;
				field-sizing: content;
				line-height: var(--font-lineheight-2);
				min-height: 32px;
				padding-block: calc((32px - 1lh) / 2);
				max-height: 180px;
				scrollbar-width: thin;
				scrollbar-color: var(--gray-9) transparent;

				&::placeholder {
					color: var(--text-muted);
				}
			}
		}

		.chat-input-actions {
			display: flex;
			justify-content: left;
			align-self: flex-start;
		}

		.chat-input-submit-btn {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 32px;
			height: 32px;
			background: var(--bg-button);
			color: white;
			border: none;
			border-radius: var(--radius-2);
			cursor: pointer;
			transition:
				background-color 0.2s var(--ease-3),
				color 0.2s var(--ease-3),
				filter 0.15s var(--ease-3);

			svg {
				width: 16px;
				height: 16px;
			}

			&:hover:not(:disabled) {
				filter: brightness(1.1);
			}

			&:disabled {
				background: var(--gray-10);
				color: var(--text-muted);
				cursor: not-allowed;
			}
		}
	}
`)

type chatInputProps struct {
	threadID    string
	isStreaming bool
}

func chatInput(props chatInputProps) Node {
	return Div(
		Class("chat-input"),
		If(props.threadID != "",
			Data("signals:thread-id", fmt.Sprintf("'%s'", props.threadID)),
		),
		Data("signals:is-streaming", fmt.Sprintf("%t", props.isStreaming)),
		Data("signals:message", "''"),
		Data("computed:can-submit", "$message.trim() !== '' && !$isStreaming"),
		Div(
			Class("chat-input-box"),
			Data("class:is-sending", "$sendingMessage"),
			Textarea(
				AutoFocus(),
				Name("message"),
				Placeholder("Type your message here..."),
				Data("indicator:sending-message", ""),
				Data("bind:message", ""),
				Data("on:keydown", `
					if (evt.key !== 'Enter' || evt.shiftKey) return;
					evt.preventDefault(); // prevent writing an enter character.
					if (!$canSubmit) return;
					@post('/chat/send');
				`),
			),
			Div(Class("chat-input-actions"),
				Button(Class("chat-input-submit-btn"),
					Data("show", "!$isStreaming"),
					Disabled(),
					Type("submit"),
					Data("attr:disabled", "!$canSubmit || $isStreaming"),
					Data("on:click", `
						@post('/chat/send')
					`),
					iconArrowUp(),
				),
				Button(Class("chat-input-submit-btn hidden"),
					Data("class:hidden", "!$isStreaming"),
					Type("submit"),
					Data("on:click", `
						@post('/chat/stop')
					`),
					iconSquare(),
				),
			),
		),
	)
}

func existingThreadChatInput(threadID string, isStreaming bool) Node {
	return chatInput(chatInputProps{
		threadID:    threadID,
		isStreaming: isStreaming,
	})
}

func newThreadChatInput() Node {
	return chatInput(chatInputProps{})
}

func PatchIsStreaming(sse *datastar.ServerSentEventGenerator, isStreaming bool) {
	patchSignals(sse, map[string]any{"isStreaming": isStreaming})
}

func ClearMessageInput(sse *datastar.ServerSentEventGenerator) {
	patchSignals(sse, map[string]any{
		"message":     "",
		"isStreaming": true,
	})
}
