package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Sidebar struct {
	Threads        []Thread
	ActiveThreadID string
	UserEmail      string
}

var _ = styles.Style(`
	.sidebar {
		width: 100%;
		font-size: var(--font-sidebar-size);
		min-height: 100vh;
		background: var(--bg-sidebar);
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--border-color);
		overflow: hidden;
	}
	.sidebar-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--size-4) var(--size-3) var(--size-3);
	}
	.app-logo {
		font-weight: var(--font-weight-4);
		letter-spacing: -0.02em;

		.accent {
			color: var(--bg-button);
		}
	}
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
	.thread-list-item {
		display: flex;
		align-items: center;
		padding: var(--size-1) var(--size-2);
		color: var(--text-secondary);
		border-radius: var(--radius-2);
		text-decoration: none;
		transition:
			background 0.12s,
			color 0.12s;

		&:hover {
			background: var(--bg-hover);
			color: var(--text-primary);
		}

		&.active {
			background: color-mix(in oklch, var(--gray-8) 18%, transparent);
			color: var(--text-primary);
		}

		&:hover .thread-archive-button {
			opacity: 1;
		}
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

func (x Sidebar) Node() Node {
	return Aside(
		Class("sidebar"),
		If(x.ActiveThreadID != "",
			Data("init", fmt.Sprintf("@get('/threads/%s/events')", x.ActiveThreadID)),
		),
		SidebarHeader(),
		NewChatButton(),
		ThreadSearch(),
		ThreadList(x.Threads, x.ActiveThreadID),
		SidebarFooter(x.UserEmail),
		ThreadListScrollRestoreScript(),
	)
}

func SidebarHeader() Node {
	return Div(Class("sidebar-header"), AppLogo())
}

func AppLogo() Node {
	return Span(
		Class("app-logo"),
		Text("afastchat"),
		Span(Class("accent"), Text(".com")),
	)
}

func NewChatButton() Node {
	return A(Href("/"), Class("new-chat-button"), iconPlus(), Text("New Chat"))
}

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
	return Div(ID("thread-items"),
		Map(threads, func(thread Thread) Node {
			return ThreadListItem(thread, thread.ID == activeThreadID)
		}),
	)
}

func ThreadListItem(thread Thread, isActive bool) Node {
	return A(
		ID("thread-"+thread.ID),
		Classes{
			"thread-list-item": true,
			"active":           isActive,
		},
		Href(fmt.Sprintf("/threads/%s", thread.ID)),
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

func SidebarFooter(email string) Node {
	return Div(
		Class("sidebar-footer"),
		UserAvatar(email),
		Form(Method("post"), Action("/logout"), LogoutButton()),
	)
}

func UserAvatar(email string) Node {
	return A(Href("/settings"), Class("sidebar-footer-link"),
		Div(Class("user-avatar"),
			Text(userInitial(email)),
		),
	)
}

func userInitial(email string) string {
	if email == "" {
		return "?"
	}
	return string([]rune(email)[0])
}

func LogoutButton() Node {
	return Button(Type("submit"), Class("logout-button"), Text("Log out"))
}
