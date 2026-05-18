package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.thread-list-item {
		display: flex;
		align-items: center;
		gap: var(--size-2);
		padding: var(--size-1) var(--size-2);
		color: var(--text-secondary);
		border-radius: var(--radius-2);
		text-decoration: none;
		transition:
			background 0.12s,
			color 0.12s,
			opacity 0.12s,
			transform 0.12s,
			box-shadow 0.12s;

		&:hover {
			background: var(--bg-hover);
			color: var(--text-primary);
		}

		&.active {
			background: color-mix(in oklch, var(--bg-button) 18%, transparent);
			color: var(--text-primary);
		}

		&:hover .thread-archive-button {
			opacity: 1;
		}
	}

	.thread-list-item-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		inline-size: 16px;
		block-size: 16px;
		flex-shrink: 0;
		color: var(--blue-4);
		opacity: 0.75;

		svg {
			inline-size: 15px;
			block-size: 15px;
		}
	}

	.thread-list-item.active .thread-list-item-icon {
		opacity: 1;
		color: var(--bg-button);
	}

	.thread-list-item-title {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
		flex: 1;
	}

	@keyframes thread-streaming-spin {
		to {
			transform: rotate(360deg);
		}
	}

	.thread-streaming-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 14px;
		height: 14px;
		flex-shrink: 0;
		color: var(--bg-button);

		svg {
			width: 14px;
			height: 14px;
			transform-origin: center;
			animation: thread-streaming-spin 0.8s linear infinite;
		}
	}

	.thread-archive-button {
		display: flex;
		opacity: 0;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: var(--size-1);
		border-radius: var(--radius-1);
		flex-shrink: 0;

		svg {
			width: 14px;
			height: 14px;
		}

		&:hover {
			color: var(--text-primary);
			background: var(--bg-hover);
		}
	}
`)

func ThreadListItem(thread Thread, isActive bool) Node {
	return A(
		ID("thread-"+thread.ID),
		Classes{
			"thread-list-item": true,
			"active":           isActive,
		},
		Attr("draggable", "true"),
		Data("on:dragstart", fmt.Sprintf(`
			$editingFolderMode = '';
			$editingFolderId = '';
			evt.dataTransfer.effectAllowed = 'move';
			evt.dataTransfer.setData('text/thread-id', '%s');
			document.body.classList.add('is-sidebar-dragging');
			el.classList.add('is-drag-source');
		`, thread.ID)),
		Data("on:dragend", `
			document.body.classList.remove('is-sidebar-dragging');
			document.querySelectorAll('.is-drop-target').forEach(node => node.classList.remove('is-drop-target'));
			document.querySelectorAll('.is-drag-source:not(.is-drop-pending)').forEach(node => node.classList.remove('is-drag-source'));
		`),
		Href(fmt.Sprintf("/threads/%s", thread.ID)),
		Span(Class("thread-list-item-icon"), iconChat()),
		ThreadListItemTitle(thread.Title),
		If(thread.IsStreaming, ThreadStreamingIndicator()),
		ThreadArchiveButton(thread.ID),
	)
}

func ThreadListItemTitle(title string) Node {
	return Span(Class("thread-list-item-title"), Text(title))
}

func ThreadStreamingIndicator() Node {
	return Span(Class("thread-streaming-indicator"), iconSpinner())
}

func ThreadArchiveButton(threadID string) Node {
	return Button(
		Type("button"),
		Class("thread-archive-button"),
		Data("on:click", fmt.Sprintf(`
			evt.preventDefault();
			evt.stopPropagation();
			@post('/threads/%s/archive');
		`, threadID)),
		iconArchive(),
	)
}
