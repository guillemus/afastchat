package views

import (
	"fmt"

	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type chatInputProps struct {
	threadID    string
	isStreaming bool
}

func chatInput(props chatInputProps) Node {
	return Div(
		Class("chat-input mx-auto w-full max-w-3xl shrink-0 px-6 pt-2 pb-5"),
		If(props.threadID != "",
			Data("signals:thread-id", fmt.Sprintf("'%s'", props.threadID)),
		),
		Data("signals:is-streaming", fmt.Sprintf("%t", props.isStreaming)),
		Data("signals:message", "''"),
		Data("computed:can-submit", "$message.trim() !== '' && !$isStreaming"),
		Div(
			Class("chat-input-box flex items-center rounded-xl border border-[var(--border-color)] bg-[var(--bg-input)] p-3"),
			Data("class:is-sending", "$sendingMessage"),
			Textarea(
				Class("field-sizing-content max-h-[180px] min-h-8 w-full resize-none border-0 bg-transparent py-[calc((32px-1lh)/2)] font-[inherit] text-[inherit] leading-[var(--font-lineheight-2)] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] [scrollbar-color:var(--gray-9)_transparent] [scrollbar-width:thin]"),
				AutoFocus(),
				Name("message"),
				Placeholder("Ask anything"),
				Data("indicator:sending-message", ""),
				Data("bind:message", ""),
				Data("on:keydown", `
					if (evt.key !== 'Enter' || evt.shiftKey) return;
					evt.preventDefault();
					if (!$canSubmit) return;
					@post('/chat/send');
				`),
			),
			Div(Class("flex justify-start self-start"),
				Button(Class("flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg border-0 bg-[var(--bg-button)] text-white transition enabled:hover:brightness-110 disabled:cursor-not-allowed disabled:bg-[var(--gray-10)] disabled:text-[var(--text-muted)]"),
					Data("show", "!$isStreaming"),
					Disabled(),
					Type("submit"),
					Data("attr:disabled", "!$canSubmit || $isStreaming"),
					Data("on:click", `
						@post('/chat/send')
					`),
					iconArrowUp("h-4 w-4"),
				),
				Button(Class("hidden flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg border-0 bg-[var(--bg-button)] text-white transition enabled:hover:brightness-110 disabled:cursor-not-allowed disabled:bg-[var(--gray-10)] disabled:text-[var(--text-muted)]"),
					Data("class:hidden", "!$isStreaming"),
					Type("submit"),
					Data("on:click", `
						@post('/chat/stop')
					`),
					iconSquare("h-4 w-4"),
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
