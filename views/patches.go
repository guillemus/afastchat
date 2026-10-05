package views

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

func PatchThreads(sse *datastar.ServerSentEventGenerator, threads []Thread, activeThreadID string) {
	patchTempl(sse, ThreadItems(threads, activeThreadID))
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread) {
	patchTempl(sse, ThreadListItem(thread, true))
}

func PatchThreadMessageList(
	sse *datastar.ServerSentEventGenerator, messages []MessageGroup, scrollMessages bool,
) {
	patchTempl(sse, MessageList(MessageThreadProps{Messages: messages}))
	if !scrollMessages {
		return
	}

	executeScript(sse, `
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

func PatchIsStreaming(sse *datastar.ServerSentEventGenerator, isStreaming bool) {
	PatchSignals(sse, map[string]any{"_isStreaming": isStreaming})
}

func ClearMessageInput(sse *datastar.ServerSentEventGenerator) {
	PatchSignals(sse, map[string]any{
		"message":      "",
		"_isStreaming": true,
	})
}

func PatchSignals(sse *datastar.ServerSentEventGenerator, v any) {
	patch, err := json.Marshal(v)
	if err != nil {
		slog.Error("marshal signals", "err", err)
		return
	}

	if err := sse.PatchSignals(patch); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("patch signals", "err", err)
	}
}

func patchTempl(sse *datastar.ServerSentEventGenerator, component templ.Component) {
	if err := sse.PatchElementTempl(component); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("patch elements", "err", err)
	}
}

func executeScript(sse *datastar.ServerSentEventGenerator, script string) {
	if err := sse.ExecuteScript(fmt.Sprintf("{ %s }", script)); err != nil {
		slog.Error("execute script", "err", err)
	}
}
