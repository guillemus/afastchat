package views

import (
	"fmt"
	"strconv"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Sidebar struct {
	Tree           SidebarTree
	ActiveThreadID string
	UserEmail      string
}

var _ = styles.Style(`
	.sidebar {
		width: 260px;
		min-width: 260px;
		max-width: 260px;
		font-size: var(--font-sidebar-size);
		flex-shrink: 0;
		min-height: 100vh;
		background: var(--bg-sidebar);
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--border-color);
		overflow: hidden;
	}
`)

func (x Sidebar) Node() Node {
	return Aside(
		Class("sidebar"),
		Data("signals:thread-id", fmt.Sprintf("'%s'", x.ActiveThreadID)),
		Data("signals:folder-name", "''"),
		Data("signals:editing-folder-id", "''"),
		Data("signals:editing-folder-mode", "''"),
		Data("signals:drag-thread-id", "''"),
		Data("signals:drag-folder-id", "''"),
		Data("signals:target-folder-id", "''"),
		If(x.ActiveThreadID != "",
			Data("init", fmt.Sprintf("@get('/threads/%s/events')", x.ActiveThreadID)),
		),
		SidebarHeader(),
		NewChatButton(),
		ThreadSearch(),
		ThreadList(x.Tree, x.ActiveThreadID),
		SidebarFooter(x.UserEmail),
		ThreadListScrollRestoreScript(),
	)
}

var _ = styles.Style(`
	.sidebar-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--size-4) var(--size-3) var(--size-3);
	}
`)

func SidebarHeader() Node {
	return Div(Class("sidebar-header"), AppLogo())
}

var _ = styles.Style(`
	.app-logo {
		font-weight: var(--font-weight-4);
		letter-spacing: -0.02em;

		.accent {
			color: var(--bg-button);
		}
	}
`)

func AppLogo() Node {
	return Span(
		Class("app-logo"),
		Text("afastchat"),
		Span(Class("accent"), Text(".com")),
	)
}

var _ = styles.Style(`
	.new-chat-button {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--size-2);
		margin: 0 var(--size-3) var(--size-3);
		padding: var(--size-2) var(--size-3);
		background: var(--bg-button);
		color: white;
		border: none;
		border-radius: var(--radius-2);
		font-weight: var(--font-weight-4);
		text-decoration: none;
		transition: filter 0.15s;

		svg {
			width: 14px;
			height: 14px;
			flex-shrink: 0;
		}

		&:hover {
			filter: brightness(1.1);
		}
	}
`)

func NewChatButton() Node {
	return A(Href("/"), Class("new-chat-button"), iconPlus(), Text("New Chat"))
}

func NewFolderButton() Node {
	return Button(
		Type("button"),
		Class("new-folder-button"),
		Title("New folder"),
		Aria("label", "New folder"),
		Data("on:click", `
			$editingFolderMode = 'create';
			$editingFolderId = '';
			$folderName = '';
			setTimeout(() => document.querySelector('#new-folder-name-input')?.focus(), 0);
		`),
		iconFolderPlus(),
	)
}

var _ = styles.Style(`
	.thread-search {
		margin: 0 var(--size-3) var(--size-2);
		position: relative;

		svg {
			position: absolute;
			left: var(--size-2);
			top: 50%;
			transform: translateY(-50%);
			color: var(--text-muted);
			width: 13px;
			height: 13px;
			pointer-events: none;
		}

		input {
			width: 100%;
			padding: var(--size-1) var(--size-2) var(--size-1)
				calc(var(--size-2) + 13px + var(--size-2));
			background: transparent;
			border: 1px solid transparent;
			border-radius: var(--radius-2);
			color: var(--text-secondary);
			font-family: inherit;
			font-size: inherit;
			outline: none;
			transition:
				border-color 0.15s,
				background 0.15s;

			&::placeholder {
				color: var(--text-muted);
			}

			&:focus {
				background: var(--bg-input);
				border-color: var(--border-color);
			}
		}
	}
`)

func ThreadSearch() Node {
	return Div(Class("thread-search"),
		iconSearch(),
		Input(
			Type("text"),
			Name("thread-search"),
			Placeholder("Search threads..."),
			Data("bind:thread-search", ""),
			Data("on:input__debounce.150ms", "@get('/threads/search')"),
		),
	)
}

var _ = styles.Style(`
	.thread-list {
		flex: 1;
		overflow-y: auto;
		padding: var(--size-1) var(--size-2);
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
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

var _ = styles.Style(`
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

	.new-folder-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		inline-size: var(--size-5);
		block-size: var(--size-5);
		border: 0;
		border-radius: var(--radius-2);
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			background 0.12s,
			color 0.12s;

		svg {
			inline-size: 16px;
			block-size: 16px;
		}

		&:hover {
			background: var(--bg-hover);
			color: var(--text-primary);
		}
	}
`)

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

func NewFolderInlineForm() Node {
	return Form(
		Class("folder-inline-form new-folder-inline-form"),
		Style("display: none"),
		Data("show", "$editingFolderMode === 'create'"),
		Data("on:submit__prevent", `
			if (!$folderName.trim()) return;
			@post('/folders/save');
			$editingFolderMode = '';
		`),
		Span(Class("folder-icon"), iconFolder()),
		Input(
			ID("new-folder-name-input"),
			Type("text"),
			Placeholder("Folder name"),
			Data("bind:folder-name", ""),
			Data("on:keydown", "if (evt.key === 'Escape') $editingFolderMode = '';"),
			Data("on:blur", "$editingFolderMode = '';"),
		),
	)
}

func FolderRenameInlineForm(folder SidebarFolder) Node {
	return Form(
		Class("folder-inline-form"),
		Style("display: none"),
		Data("show", fmt.Sprintf("$editingFolderId === '%s'", folder.ID)),
		Data("on:click", "evt.stopPropagation();"),
		Data("on:submit__prevent", `
			if (!$folderName.trim()) return;
			@post('/folders/save');
			$editingFolderMode = '';
			$editingFolderId = '';
		`),
		Input(
			ID("folder-name-input-"+folder.ID),
			Type("text"),
			Data("bind:folder-name", ""),
			Data("on:blur", `
				$editingFolderMode = '';
				$editingFolderId = '';
			`),
			Data("on:keydown", `
				if (evt.key !== 'Escape') return;
				$editingFolderMode = '';
				$editingFolderId = '';
			`),
		),
	)
}

func FolderItem(folder SidebarFolder, activeThreadID string) Node {
	containsActive := folderContainsActiveThread(folder, activeThreadID)
	return Details(
		ID("folder-"+folder.ID),
		Class("folder-item"),
		If(!containsActive, Data("preserve-attr", "open")),
		If(containsActive, Attr("open")),
		Data("on:dragover", "evt.preventDefault();"),
		Data("on:dragenter", "el.classList.add('is-drop-target');"),
		Data("on:dragleave", "if (!el.contains(evt.relatedTarget)) el.classList.remove('is-drop-target');"),
		Data("on:drop", fmt.Sprintf(`
			evt.stopPropagation();
			el.classList.remove('is-drop-target');
			document.body.classList.remove('is-sidebar-dragging');
			document.querySelectorAll('.is-drop-target').forEach(node => node.classList.remove('is-drop-target'));
			let threadID = evt.dataTransfer.getData('text/thread-id');
			let folderID = evt.dataTransfer.getData('text/folder-id');
			$targetFolderId = '%s';
			if (threadID) {
				$dragThreadId = threadID;
				document.querySelector('.is-drag-source')?.classList.add('is-drop-pending');
				@post('/threads/move');
				return;
			}
			if (folderID) {
				$dragFolderId = folderID;
				if ($dragFolderId === $targetFolderId) return;
				document.querySelector('.is-drag-source')?.classList.add('is-drop-pending');
				@post('/folders/move');
			}
		`, folder.ID)),
		Summary(
			Class("folder-summary"),
			Attr("draggable", "true"),
			Data("on:dragstart", fmt.Sprintf(`
				evt.stopPropagation();
				$editingFolderMode = '';
				$editingFolderId = '';
				evt.dataTransfer.effectAllowed = 'move';
				evt.dataTransfer.setData('text/folder-id', '%s');
				document.body.classList.add('is-sidebar-dragging');
				el.closest('.folder-item').classList.add('is-drag-source');
			`, folder.ID)),
			Data("on:dragend", `
				document.body.classList.remove('is-sidebar-dragging');
				document.querySelectorAll('.is-drop-target').forEach(node => node.classList.remove('is-drop-target'));
				document.querySelectorAll('.is-drag-source:not(.is-drop-pending)').forEach(node => node.classList.remove('is-drag-source'));
			`),
			Data("on:dblclick", fmt.Sprintf(`
				evt.preventDefault();
				evt.stopPropagation();
				$editingFolderMode = 'rename';
				$editingFolderId = '%s';
				$folderName = %s;
				setTimeout(() => document.querySelector('#folder-name-input-%s')?.focus(), 0);
			`, folder.ID, strconv.Quote(folder.Name), folder.ID)),
			Span(Class("folder-icon folder-icon-closed"), iconFolder()),
			Span(Class("folder-icon folder-icon-open"), iconFolderOpen()),
			Span(Class("folder-name"), Data("show", fmt.Sprintf("$editingFolderId !== '%s'", folder.ID)), Text(folder.Name)),
			FolderRenameInlineForm(folder),
			FolderDeleteButton(folder),
		),
		Div(
			Class("folder-children"),
			Map(folder.Folders, func(child SidebarFolder) Node {
				return FolderItem(child, activeThreadID)
			}),
			Map(folder.Threads, func(thread Thread) Node {
				return ThreadListItem(thread, thread.ID == activeThreadID)
			}),
		),
	)
}

func folderContainsActiveThread(folder SidebarFolder, activeThreadID string) bool {
	for _, thread := range folder.Threads {
		if thread.ID == activeThreadID {
			return true
		}
	}
	for _, child := range folder.Folders {
		if folderContainsActiveThread(child, activeThreadID) {
			return true
		}
	}
	return false
}

var _ = styles.Style(`
	.sidebar-items {
		min-height: 100%;
		border-radius: var(--radius-2);
		transition: background 0.12s, box-shadow 0.12s;
	}

	.sidebar-items.is-drop-target {
		background: color-mix(in oklch, var(--bg-button) 8%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in oklch, var(--bg-button) 35%, transparent);
	}

	.folder-item {
		margin-bottom: var(--size-1);
		border-radius: var(--radius-2);
		transition: opacity 0.12s, transform 0.12s;
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

	.folder-item.is-drop-target > .folder-summary {
		background: color-mix(in oklch, var(--bg-button) 18%, var(--bg-hover));
		box-shadow:
			inset 0 0 0 1px var(--bg-button),
			0 0 0 3px color-mix(in oklch, var(--bg-button) 16%, transparent);
		color: var(--text-primary);
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

	.folder-summary {
		display: flex;
		align-items: center;
		gap: var(--size-2);
		min-block-size: var(--size-7);
		padding: var(--size-1) var(--size-2);
		color: var(--gray-3);
		border-radius: var(--radius-2);
		cursor: pointer;
		list-style: none;
		font-weight: var(--font-weight-5);

		&::-webkit-details-marker {
			display: none;
		}

		&:hover {
			background: color-mix(in oklch, var(--bg-hover) 85%, var(--orange-9));
			color: var(--text-primary);
		}

		&:hover .folder-delete-button {
			opacity: 1;
		}
	}

	.folder-inline-form {
		display: flex;
		align-items: center;
		gap: var(--size-2);
		min-width: 0;
		flex: 1;
	}

	.new-folder-inline-form {
		padding: var(--size-1) var(--size-2);
		color: var(--gray-3);
	}

	.folder-inline-form input {
		min-width: 0;
		width: 100%;
		block-size: var(--size-5);
		border: 0;
		border-radius: var(--radius-2);
		background: var(--bg-input);
		box-shadow: inset 0 0 0 1px var(--bg-button);
		color: var(--text-primary);
		font: inherit;
		font-weight: var(--font-weight-5);
		line-height: var(--font-lineheight-1);
		padding: 0 var(--size-2);
		outline: none;
	}

	.folder-inline-form input:focus {
		box-shadow: 0 0 0 2px color-mix(in oklch, var(--bg-button) 18%, transparent);
	}

	.folder-icon,
	.thread-list-item-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		inline-size: 16px;
		block-size: 16px;
		flex-shrink: 0;

		svg {
			inline-size: 15px;
			block-size: 15px;
		}
	}

	.folder-icon {
		color: var(--orange-4);
		filter: drop-shadow(0 0 6px color-mix(in oklch, var(--orange-5) 20%, transparent));
	}

	.folder-icon-open {
		display: none;
	}

	.folder-item[open] > .folder-summary .folder-icon-closed {
		display: none;
	}

	.folder-item[open] > .folder-summary .folder-icon-open {
		display: inline-flex;
	}

	.folder-name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
		flex: 1;
		padding-inline: var(--size-2);
	}

	.folder-children {
		margin-left: var(--size-2);
		padding-left: var(--size-3);
		border-left: 1px solid color-mix(in oklch, var(--border-color) 75%, transparent);
	}

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
			box-shadow: inset 2px 0 0 var(--bg-button);
		}

		.thread-list-item-icon {
			color: var(--blue-4);
			opacity: 0.75;
		}

		&.active .thread-list-item-icon {
			opacity: 1;
			color: var(--bg-button);
		}

		&:hover .thread-archive-button {
			opacity: 1;
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

var _ = styles.Style(`
	.thread-list-item-title {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
		flex: 1;
	}
`)

func ThreadListItemTitle(title string) Node {
	return Span(Class("thread-list-item-title"), Text(title))
}

var _ = styles.Style(`
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
`)

func ThreadStreamingIndicator() Node {
	return Span(Class("thread-streaming-indicator"), iconSpinner())
}

var _ = styles.Style(`
	.folder-delete-button,
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

func FolderDeleteButton(folder SidebarFolder) Node {
	message := fmt.Sprintf("Delete “%s”? This deletes %d folders and %d chats. This cannot be undone.", folder.Name, folder.DeleteFolderCount, folder.DeleteThreadCount)
	return Button(
		Type("button"),
		Class("folder-delete-button"),
		Data("on:click", fmt.Sprintf(`
			evt.preventDefault();
			evt.stopPropagation();
			if (!confirm(%s)) return;
			@post('/folders/%s/delete');
		`, strconv.Quote(message), folder.ID)),
		iconArchive(),
	)
}

var _ = styles.Style(`
	.sidebar-footer {
		padding: var(--size-3);
		border-top: 1px solid var(--border-color);
		display: flex;
		align-items: center;
		gap: var(--size-2);

		form {
			margin-left: auto;
		}
	}
`)

func SidebarFooter(email string) Node {
	return Div(
		Class("sidebar-footer"),
		UserAvatar(email),
		Form(Method("post"), Action("/logout"), LogoutButton()),
	)
}

var _ = styles.Style(`
	.user-avatar {
		width: 26px;
		height: 26px;
		border-radius: var(--radius-round);
		background: var(--bg-button);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: var(--font-weight-4);
		color: white;
		flex-shrink: 0;
	}
`)

func UserAvatar(email string) Node {
	return Div(Class("user-avatar"), Text(userInitial(email)))
}

func userInitial(email string) string {
	if email == "" {
		return "?"
	}
	return string([]rune(email)[0])
}

var _ = styles.Style(`
	.logout-button {
		background: transparent;
		border: 1px solid var(--border-color);
		color: var(--text-secondary);
		border-radius: var(--radius-2);
		padding: 0 var(--size-2);
		min-height: 32px;
		font-family: inherit;
		font-size: inherit;
		cursor: pointer;

		&:hover {
			color: var(--text-primary);
			background: var(--bg-hover);
		}
	}
`)

func LogoutButton() Node {
	return Button(Type("submit"), Class("logout-button"), Text("Log out"))
}
