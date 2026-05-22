package views

import (
	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.toast-region {
		position: fixed;
		right: var(--size-5);
		bottom: var(--size-5);
		z-index: 10;
	}
	.toast {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: var(--size-1) var(--size-3);
		min-width: 280px;
		max-width: 420px;
		border: 1px solid var(--red-8);
		border-radius: var(--radius-3);
		background: var(--gray-11);
		color: var(--text-primary);
		box-shadow: var(--shadow-3);
		padding: var(--size-3) var(--size-4);
		opacity: 1;
		transition: opacity 100ms var(--ease-3);
	}
	.toast.is-hiding {
		opacity: 0;
	}
	.toast-title {
		font-weight: var(--font-weight-6);
	}
	.toast-message {
		grid-column: 1 / -1;
		color: var(--text-secondary);
		line-height: var(--font-lineheight-2);
	}
	.toast-action {
		grid-column: 1 / -1;
		color: var(--text-primary);
		font-weight: var(--font-weight-5);
	}
	.toast-close {
		border: 0;
		background: transparent;
		color: var(--text-secondary);
		font: inherit;
		cursor: pointer;
		line-height: 1;
	}
`)

func ToastRegion() Node {
	return Div(ID("toast-region"), Class("toast-region"))
}

type ToastProps struct {
	Title   string
	Message string
	Action  *ToastAction
}

type ToastAction struct {
	Label string
	Href  string
}

func Toast(data ToastProps) Node {
	var action Node
	if data.Action != nil {
		action = A(Class("toast-action"), Href(data.Action.Href), Text(data.Action.Label))
	}

	return Div(ID("toast"), Class("toast"),
		Div(Class("toast-title"), Text(data.Title)),
		Button(Type("button"),
			Class("toast-close"),
			Aria("label", "Close"),
			Data("on:click", `document.getElementById('toast')?.remove()`), Text("×")),
		Div(Class("toast-message"), Text(data.Message)),
		action,
	)
}

func PatchErrorToast(sse *datastar.ServerSentEventGenerator, title, message string) {
	PatchToast(sse, ToastProps{Title: title, Message: message})
}

func PatchToast(sse *datastar.ServerSentEventGenerator, data ToastProps) {
	ssePatch(sse, Div(ID("toast-region"), Class("toast-region"), Toast(data)))
	timeout := "3000"
	if data.Action != nil {
		timeout = "6000"
	}
	sseExecJS(sse, `setTimeout(() => {
		const toast = document.getElementById('toast');
		if (!toast) return;
		toast.classList.add('is-hiding');
		setTimeout(() => toast.remove(), 100);
	}, `+timeout+`)`)
}
