package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.thread-list {
		flex: 1;
		overflow-y: auto;
		padding: var(--size-1) var(--size-2);
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
	}

	.thread-section-label {
		font-weight: var(--font-weight-4);
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--text-muted);
		padding: var(--size-1) var(--size-2) var(--size-1);
		margin-bottom: var(--size-1);
	}
`)

func ThreadList(threads []Thread, activeThreadID string) Node {
	return Div(
		Class("thread-list"),
		Data("on:scroll__debounce.10ms", `
			setSessionVal('thread_list_scroll', String(el.scrollTop));
		`),
		ThreadSectionLabel("Recent"),
		ThreadItems(threads, activeThreadID),
	)
}

func ThreadListScrollRestoreScript() Node {
	// Keep this after all sidebar layout-affecting nodes, especially SidebarFooter.
	// ScrollTop is clamped to the current max scroll; restoring before the footer is parsed
	// uses a too-tall thread list, then the footer shrinks it and leaves scroll above bottom.
	return Script(Raw(`{
		let el = document.querySelector('.thread-list')
		let value = getSessionVal('thread_list_scroll');
		if (value) {
			el.scrollTop = Number.parseFloat(value) || 0;
		}
	}`))
}

func ThreadSectionLabel(label string) Node {
	return Div(ID("thread-section-recent"), Class("thread-section-label"), Text(label))
}

func ThreadItems(threads []Thread, activeThreadID string) Node {
	return Div(
		ID("thread-items"),
		Map(threads, func(thread Thread) Node {
			return ThreadListItem(thread, thread.ID == activeThreadID)
		}),
	)
}
