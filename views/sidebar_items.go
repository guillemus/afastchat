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
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-weight: var(--font-weight-4);
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--text-muted);
		padding: var(--size-1) var(--size-2) var(--size-1);
		margin-bottom: var(--size-1);
	}

	.sidebar-items {
		min-height: 100%;
		border-radius: var(--radius-2);
		transition: background 0.12s, box-shadow 0.12s;
	}

	.sidebar-items.is-drop-target {
		background: color-mix(in oklch, var(--bg-button) 8%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in oklch, var(--bg-button) 35%, transparent);
	}

	.folder-item.is-drag-source,
	.thread-list-item.is-drag-source {
		opacity: 0.38;
		transform: scale(0.98);
	}

	.folder-item.is-drop-pending > .folder-summary,
	.thread-list-item.is-drop-pending {
		opacity: 0.22;
		transform: scale(0.96);
		filter: grayscale(1);
		background: repeating-linear-gradient(
			-45deg,
			color-mix(in oklch, var(--bg-sidebar) 70%, black),
			color-mix(in oklch, var(--bg-sidebar) 70%, black) var(--size-1),
			color-mix(in oklch, var(--bg-hover) 45%, transparent) var(--size-1),
			color-mix(in oklch, var(--bg-hover) 45%, transparent) var(--size-2)
		);
		box-shadow: inset 0 0 0 1px color-mix(in oklch, var(--bg-button) 45%, transparent);
		pointer-events: none;
	}

	body.is-sidebar-dragging .folder-summary,
	body.is-sidebar-dragging .thread-list-item {
		cursor: grabbing;
	}

	body.is-sidebar-dragging .folder-delete-button,
	body.is-sidebar-dragging .thread-archive-button {
		opacity: 0;
		pointer-events: none;
	}
`)

func ThreadList(tree SidebarTree, activeThreadID string) Node {
	return Div(
		Class("thread-list"),
		Data("on:scroll__debounce.10ms", `
			setSessionVal('thread_list_scroll', String(el.scrollTop));
		`),
		ThreadSectionLabel("Recent"),
		SidebarItems(tree, activeThreadID),
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
	return Div(ID("thread-section-recent"), Class("thread-section-label"), Span(Text(label)), NewFolderButton())
}

func SidebarItems(tree SidebarTree, activeThreadID string) Node {
	return Div(
		ID("thread-items"),
		Class("sidebar-items"),
		Data("on:pointerdown", `
			if (evt.target.closest('.folder-inline-form')) return;
			$editingFolderMode = '';
			$editingFolderId = '';
		`),
		Data("on:dragover", "evt.preventDefault();"),
		Data("on:dragenter", "if (evt.target === el) el.classList.add('is-drop-target');"),
		Data("on:dragleave", "if (evt.target === el) el.classList.remove('is-drop-target');"),
		Data("on:drop", `
			el.classList.remove('is-drop-target');
			document.body.classList.remove('is-sidebar-dragging');
			document.querySelectorAll('.is-drop-target').forEach(node => node.classList.remove('is-drop-target'));
			let threadID = evt.dataTransfer.getData('text/thread-id');
			let folderID = evt.dataTransfer.getData('text/folder-id');
			$targetFolderId = '';
			if (threadID) {
				$dragThreadId = threadID;
				document.querySelector('.is-drag-source')?.classList.add('is-drop-pending');
				@post('/threads/move');
				return;
			}
			if (folderID) {
				$dragFolderId = folderID;
				document.querySelector('.is-drag-source')?.classList.add('is-drop-pending');
				@post('/folders/move');
			}
		`),
		NewFolderInlineForm(),
		Map(tree.Folders, func(folder SidebarFolder) Node {
			return FolderItem(folder, activeThreadID)
		}),
		Map(tree.Threads, func(thread Thread) Node {
			return ThreadListItem(thread, thread.ID == activeThreadID)
		}),
	)
}
