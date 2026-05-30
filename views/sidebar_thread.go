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
			"thread-list-item group flex items-center gap-2 rounded-lg px-2 py-1 text-neutral-400 no-underline transition-colors duration-100 hover:bg-neutral-900 hover:text-neutral-100": true,
			"bg-[color-mix(in_oklch,#f97316_18%,transparent)] text-neutral-100": isActive,
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
	return Span(Class("flex h-3.5 w-3.5 shrink-0 items-center justify-center text-orange-500"), iconSpinner("h-3.5 w-3.5 origin-center animate-[thread-streaming-spin_0.8s_linear_infinite]"))
}

func ThreadArchiveButton(threadID string) Node {
	return Button(
		Type("button"),
		Class("flex shrink-0 cursor-pointer items-center justify-center rounded p-1 text-neutral-500 opacity-0 group-hover:opacity-100 hover:bg-neutral-900 hover:text-neutral-100"),
		Data("on:click", fmt.Sprintf(`
			evt.preventDefault();
			evt.stopPropagation();
			@post('/threads/%s/archive');
		`, threadID)),
		iconArchive("h-3.5 w-3.5"),
	)
}
