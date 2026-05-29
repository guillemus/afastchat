package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func ThreadListItem(thread Thread, isActive bool) Node {
	return A(
		ID("thread-"+thread.ID),
		Classes{
			"thread-list-item group flex items-center gap-2 rounded-lg px-2 py-1 text-[var(--text-secondary)] no-underline transition-colors duration-100 hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]": true,
			"bg-[color-mix(in_oklch,var(--bg-button)_18%,transparent)] text-[var(--text-primary)]":                                                                                                                     isActive,
		},
		Href(fmt.Sprintf("/threads/%s", thread.ID)),
		ThreadListItemTitle(thread.Title),
		If(thread.IsStreaming, ThreadStreamingIndicator()),
		ThreadArchiveButton(thread.ID),
	)
}

func ThreadListItemTitle(title string) Node {
	return Span(Class("min-w-0 flex-1 truncate"), Text(title))
}

func ThreadStreamingIndicator() Node {
	return Span(Class("flex h-3.5 w-3.5 shrink-0 items-center justify-center text-[var(--bg-button)] [&_svg]:h-3.5 [&_svg]:w-3.5 [&_svg]:origin-center [&_svg]:animate-[thread-streaming-spin_0.8s_linear_infinite]"), iconSpinner())
}

func ThreadArchiveButton(threadID string) Node {
	return Button(
		Type("button"),
		Class("flex shrink-0 cursor-pointer items-center justify-center rounded p-1 text-[var(--text-muted)] opacity-0 group-hover:opacity-100 hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)] [&_svg]:h-3.5 [&_svg]:w-3.5"),
		Data("on:click", fmt.Sprintf(`
			evt.preventDefault();
			evt.stopPropagation();
			@post('/threads/%s/archive');
		`, threadID)),
		iconArchive(),
	)
}
