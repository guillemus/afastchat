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
			Class("chat-input-box flex items-center rounded-xl border border-[var(--border-color)] bg-[var(--bg-input)] p-3 [&_textarea]:field-sizing-content [&_textarea]:max-h-[180px] [&_textarea]:min-h-8 [&_textarea]:w-full [&_textarea]:resize-none [&_textarea]:border-0 [&_textarea]:bg-transparent [&_textarea]:py-[calc((32px-1lh)/2)] [&_textarea]:font-[inherit] [&_textarea]:text-[inherit] [&_textarea]:leading-[var(--font-lineheight-2)] [&_textarea]:text-[var(--text-primary)] [&_textarea]:outline-none [&_textarea]:[scrollbar-color:var(--gray-9)_transparent] [&_textarea]:[scrollbar-width:thin] [&_textarea::placeholder]:text-[var(--text-muted)]"),
			Data("class:is-sending", "$sendingMessage"),
			Textarea(
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
				Button(Class("flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg border-0 bg-[var(--bg-button)] text-white transition enabled:hover:brightness-110 disabled:cursor-not-allowed disabled:bg-[var(--gray-10)] disabled:text-[var(--text-muted)] [&_svg]:h-4 [&_svg]:w-4"),
					Data("show", "!$isStreaming"),
					Disabled(),
					Type("submit"),
					Data("attr:disabled", "!$canSubmit || $isStreaming"),
					Data("on:click", `
						@post('/chat/send')
					`),
					iconArrowUp(),
				),
				Button(Class("hidden flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg border-0 bg-[var(--bg-button)] text-white transition enabled:hover:brightness-110 disabled:cursor-not-allowed disabled:bg-[var(--gray-10)] disabled:text-[var(--text-muted)] [&_svg]:h-4 [&_svg]:w-4"),
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
