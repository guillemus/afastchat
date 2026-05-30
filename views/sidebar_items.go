package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ThreadList(threads []Thread, activeThreadID string) Node {
	return Div(
		Class("thread-list flex-1 overflow-y-auto overscroll-contain px-2 py-1 [scrollbar-color:theme(colors.neutral.700)_transparent] [scrollbar-width:thin]"),
		Data("on:scroll__debounce.10ms", `
			setSessionVal('thread_list_scroll', String(el.scrollTop));
		`),
		ThreadItems(threads, activeThreadID),
	)
}

// Keep this after all sidebar layout-affecting nodes, especially SidebarFooter.
// ScrollTop is clamped to the current max scroll; restoring before the footer is parsed
// uses a too-tall thread list, then the footer shrinks it and leaves scroll above bottom.
var ThreadListScrollRestoreScript = Static(Script(Raw(`{
	let el = document.querySelector('.thread-list')
	let value = getSessionVal('thread_list_scroll');
	if (value) {
		el.scrollTop = Number.parseFloat(value) || 0;
	}
}`)))

func ThreadSectionLabel(label string) Node {
	return Div(ID("thread-section-recent"), Class("mb-1 px-2 py-1 text-xs font-normal uppercase tracking-[0.08em] text-neutral-500"), Text(label))
}

func ThreadItems(threads []Thread, activeThreadID string) Node {
	return Div(
		ID("thread-items"),
		Map(threads, func(thread Thread) Node {
			return ThreadListItem(thread, thread.ID == activeThreadID)
		}),
	)
}
