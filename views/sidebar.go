package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SidebarProps struct {
	Threads        []Thread
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

func Sidebar(x SidebarProps) Node {
	return Aside(
		Class("sidebar"),
		Data("signals:thread-id", fmt.Sprintf("'%s'", x.ActiveThreadID)),
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

func SidebarFooter(email string) Node {
	return Div(
		Class("sidebar-footer"),
		UserAvatar(email),
		Form(Method("post"), Action("/logout"), LogoutButton()),
	)
}

func UserAvatar(email string) Node {
	return Div(Class("user-avatar"), Text(userInitial(email)))
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
