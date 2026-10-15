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
	patchTempl(sse, ThreadItems(threads, activeThreadID), datastar.WithModeReplace())
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread, isActive bool) {
	patchTempl(sse, ThreadListItem(thread, isActive))
}

// PatchThreadPage is sent whenever a thread events stream connects: on page
// load, on reconnect, on a sidebar switch and on back/forward. Only the last
// two change the thread on screen, and only then should the messages scroll
// to the bottom, the draft be cleared and the input be focused.
//
// The first script records whether this is a switch by checking if the
// thread's sidebar item was already current before the patch. It must run
// before the patch, because afterwards the item is always current. The URL
// can't answer it either: on back/forward the browser has already changed it.
//
// The last script pushes the URL only when it differs, so it is a no-op on
// load, reconnect and back/forward.
func PatchThreadPage(sse *datastar.ServerSentEventGenerator, data ThreadData) {
	path := "/threads/" + data.Thread.ID
	executeScript(sse, fmt.Sprintf(
		`window.threadSwitched = document.getElementById(%q)?.getAttribute('aria-current') !== 'page'`,
		"thread-"+data.Thread.ID,
	))
	patchTempl(sse, ThreadPage(data, false))
	PatchIsStreaming(sse, data.Streaming)
	executeScript(sse, fmt.Sprintf(`
		if (location.pathname !== %[1]q) history.pushState(null, '', %[1]q);
		if (window.threadSwitched) {
			const messages = document.getElementById('messages');
			const pinned = messages?.querySelector('[data-scroll-pin]')?.previousElementSibling;
			if (pinned) {
				pinned.scrollIntoView({ block: 'start' });
			} else if (messages) {
				messages.scrollTop = messages.scrollHeight;
			}
			const input = document.querySelector('textarea[name=message]');
			if (input) {
				input.value = '';
				input.dispatchEvent(new Event('input', { bubbles: true }));
				if (!document.activeElement?.closest('[data-editing]')) input.focus();
			}
		}
	`, path))
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
			latestUserMessage.scrollIntoView({ block: 'start', behavior: 'smooth' });
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
		"message":       "",
		"images":        []any{},
		"_imagesStatus": "",
		"_isStreaming":  true,
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

func patchTempl(sse *datastar.ServerSentEventGenerator, component templ.Component, opts ...datastar.PatchElementOption) {
	if err := sse.PatchElementTempl(component, opts...); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("patch elements", "err", err)
	}
}

func executeScript(sse *datastar.ServerSentEventGenerator, script string) {
	if err := sse.ExecuteScript(fmt.Sprintf("{ %s }", script)); err != nil {
		slog.Error("execute script", "err", err)
	}
}
