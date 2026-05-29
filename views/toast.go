package views

import (
	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ToastRegion() Node {
	return Div(ID("toast-region"), Class("toast-region fixed right-5 bottom-5 z-10"))
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
		action = A(Class("col-span-full font-medium text-[var(--text-primary)]"), Href(data.Action.Href), Text(data.Action.Label))
	}

	return Div(ID("toast"), Class("grid min-w-[280px] max-w-[420px] grid-cols-[1fr_auto] gap-x-3 gap-y-1 rounded-xl border border-[var(--red-8)] bg-[var(--gray-11)] px-4 py-3 text-[var(--text-primary)] opacity-100 shadow-[var(--shadow-3)] transition-opacity duration-100"),
		Div(Class("font-semibold"), Text(data.Title)),
		Button(Type("button"),
			Class("cursor-pointer border-0 bg-transparent font-[inherit] leading-none text-[var(--text-secondary)]"),
			Aria("label", "Close"),
			Data("on:click", `document.getElementById('toast')?.remove()`), Text("×")),
		Div(Class("col-span-full leading-6 text-[var(--text-secondary)]"), Text(data.Message)),
		action,
	)
}

func PatchErrorToast(sse *datastar.ServerSentEventGenerator, title, message string) {
	PatchToast(sse, ToastProps{Title: title, Message: message})
}

func PatchToast(sse *datastar.ServerSentEventGenerator, data ToastProps) {
	ssePatch(sse, Div(ID("toast-region"), Class("toast-region fixed right-5 bottom-5 z-10"), Toast(data)))
	timeout := "3000"
	if data.Action != nil {
		timeout = "6000"
	}
	sseExecJS(sse, `setTimeout(() => {
		const toast = document.getElementById('toast');
		if (!toast) return;
		toast.classList.add('opacity-0');
		setTimeout(() => toast.remove(), 100);
	}, `+timeout+`)`)
}
