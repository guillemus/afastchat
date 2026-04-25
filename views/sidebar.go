package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Sidebar struct {
	Threads         []Thread
	ActiveThreadID  string
	DoScrollSidebar bool
	UserEmail       string
}

var _ = styler.Style(`
	.sidebar {
		width: 260px;
		min-width: 260px;
		max-width: 260px;
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
		If(x.ActiveThreadID != "",
			Data("init", fmt.Sprintf("@get('/threads/%s/events')", x.ActiveThreadID)),
		),
		SidebarHeader(),
		NewChatButton(),
		ThreadSearch(),
		ThreadList(x.Threads, x.ActiveThreadID, x.DoScrollSidebar),
		SidebarFooter(x.UserEmail),
	)
}

var _ = styler.Style(`
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

var _ = styler.Style(`
	.app-logo {
		font-size: var(--font-size-1);
		font-weight: var(--font-weight-7);
		color: var(--text-primary);
		letter-spacing: -0.02em;

		.accent {
			color: var(--bg-button);
			font-weight: var(--font-weight-4);
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

var _ = styler.Style(`
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
		cursor: pointer;
		font-family: inherit;
		font-size: var(--font-size-0);
		font-weight: var(--font-weight-5);
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
	return A(Href("/app"), Class("new-chat-button"), iconPlus(), Text("New Chat"))
}

var _ = styler.Style(`
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
			font-size: var(--font-size-0);
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

var _ = styler.Style(`
	.thread-list {
		flex: 1;
		overflow-y: auto;
		padding: var(--size-1) var(--size-2);
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
	}
`)

func ThreadList(threads []Thread, activeThreadID string, doScrollSidebar bool) Node {
	return Div(
		Classes{
			"thread-list": true,
			"invisible":   doScrollSidebar,
		},
		If(doScrollSidebar, Group{
			Data("init", `
				const value = getSessionVal('thread_list_scroll');
				if (value) {
					el.scrollTop = Number.parseInt(value, 10) || 0;
				}

				el.classList.remove('invisible');
			`),
		}),
		Data("on:scroll__debounce.10ms", `
			let val = String(Math.round(el.scrollTop));
			setSessionVal('thread_list_scroll', val);
		`),
		ThreadSectionLabel("Recent"),
		ThreadItems(threads, activeThreadID),
	)
}

var _ = styler.Style(`
	.thread-section-label {
		font-size: 10px;
		font-weight: var(--font-weight-6);
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--text-muted);
		padding: var(--size-1) var(--size-2) var(--size-1);
		margin-bottom: var(--size-1);
	}
`)

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

var _ = styler.Style(`
	.thread-list-item {
		display: flex;
		align-items: center;
		padding: var(--size-1) var(--size-2);
		font-size: var(--font-size-0);
		color: var(--text-secondary);
		border-radius: var(--radius-2);
		cursor: pointer;
		text-decoration: none;
		position: relative;
		transition:
			background 0.12s,
			color 0.12s;
		border-left: 2px solid transparent;

		&:hover {
			background: var(--bg-hover);
			color: var(--text-primary);
			border-left-color: color-mix(in oklch, var(--bg-button) 50%, transparent);
		}

		&.active {
			background: color-mix(in oklch, var(--bg-button) 12%, transparent);
			color: var(--text-primary);
			border-left-color: var(--bg-button);
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
		Href(fmt.Sprintf("/threads/%s", thread.ID)),
		ThreadListItemTitle(thread.Title),
		If(thread.IsStreaming, ThreadStreamingIndicator()),
		ThreadArchiveButton(thread.ID),
	)
}

var _ = styler.Style(`
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

var _ = styler.Style(`
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

var _ = styler.Style(`
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

var _ = styler.Style(`
	.sidebar-footer {
		padding: var(--size-3);
		border-top: 1px solid var(--border-color);
		display: flex;
		align-items: center;
		gap: var(--size-2);
	}
`)

func SidebarFooter(email string) Node {
	return Div(
		Class("sidebar-footer"),
		UserAvatar(email),
		Span(Class("sidebar-user-email"), Text(email)),
		Form(Method("post"), Action("/logout"), LogoutButton()),
	)
}

var _ = styler.Style(`
	.user-avatar {
		width: 26px;
		height: 26px;
		border-radius: var(--radius-round);
		background: var(--bg-button);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 11px;
		font-weight: var(--font-weight-7);
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

var _ = styler.Style(`
	.sidebar-user-email {
		font-size: var(--font-size-0);
		color: var(--text-secondary);
		min-width: 0;
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
`)

var _ = styler.Style(`
	.logout-button {
		background: transparent;
		border: 1px solid var(--border-color);
		color: var(--text-secondary);
		border-radius: var(--radius-2);
		padding: 0 var(--size-2);
		min-height: 32px;
		font-family: inherit;
		font-size: var(--font-size-00);
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
