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

func PatchIsStreaming(sse *datastar.ServerSentEventGenerator, isStreaming bool) {
	patchSignals(sse, map[string]any{"isStreaming": isStreaming})
}

func ClearMessageInput(sse *datastar.ServerSentEventGenerator) {
	patchSignals(sse, map[string]any{
		"message":     "",
		"isStreaming": true,
	})
}

func chatInput(props chatInputProps) Node {
	return Div(
		Class("chat-input"),
		If(props.threadID != "",
			Data("signals:thread-id", fmt.Sprintf("'%s'", props.threadID)),
		),
		Data("signals:is-streaming", fmt.Sprintf("%t", props.isStreaming)),
		Data("signals:message", "''"),
		Data("computed:can-submit", "$message.trim() !== ''"),
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
					if (!$canSubmit) return;
					evt.preventDefault(); // prevent writing an enter character.
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
