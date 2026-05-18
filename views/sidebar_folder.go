package views

import (
	"fmt"
	"strconv"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
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

	.folder-item {
		margin-bottom: var(--size-1);
		border-radius: var(--radius-2);
		transition: opacity 0.12s, transform 0.12s;
	}

	.folder-item.is-drop-target > .folder-summary {
		background: color-mix(in oklch, var(--bg-button) 18%, var(--bg-hover));
		box-shadow:
			inset 0 0 0 1px var(--bg-button),
			0 0 0 3px color-mix(in oklch, var(--bg-button) 16%, transparent);
		color: var(--text-primary);
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

	.folder-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		inline-size: 16px;
		block-size: 16px;
		flex-shrink: 0;
		color: var(--orange-4);
		filter: drop-shadow(0 0 6px color-mix(in oklch, var(--orange-5) 20%, transparent));

		svg {
			inline-size: 15px;
			block-size: 15px;
		}
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

	.folder-delete-button {
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
