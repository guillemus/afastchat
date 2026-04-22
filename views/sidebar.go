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
}

func (x Sidebar) Node() Node {
	return Aside(
		Class("sidebar"),
		If(x.ActiveThreadID != "",
			Data("init", fmt.Sprintf("@get('/threads/%s/events')", x.ActiveThreadID)),
		),

		Div(Class("sidebar-header"),
			Span(
				Class("logo"),
				Text("afastchat"),
				Span(Class("logo-accent"), Text(".com")),
			),
		),
		A(Href("/app"), Class("new-chat-btn"),
			iconPlus(),
			Text("New Chat"),
		),
		Div(Class("search-box"),
			iconSearch(),
			Input(
				Type("text"),
				Name("thread-search"),
				Placeholder("Search threads..."),
				Data("bind:thread-search", ""),
				Data("on:input__debounce.150ms", "@get('/threads/search')"),
			),
		),
		Div(
			Classes{
				"thread-list": true,
				"invisible":   x.DoScrollSidebar,
			},
			If(x.DoScrollSidebar, Group{
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

			Div(
				ID("thread-section-recent"),
				Class("thread-section-label"),
				Text("Recent"),
			),
			ThreadList(x.Threads, x.ActiveThreadID),
		),
		Div(
			Class("sidebar-footer"),
			Div(Class("avatar"), Text("G")),
			Span(Class("sidebar-footer-name"), Text("Guillem")),
		),
	)
}

func ThreadList(threads []Thread, activeThreadID string) Node {
	return Div(ID("thread-items"),
		Map(threads, func(thread Thread) Node {
			return threadListItem(thread, thread.ID == activeThreadID)
		}),
	)
}

func threadListItem(thread Thread, isActive bool) Node {
	className := "thread-item"
	if isActive {
		className += " active"
	}

	return A(
		ID("thread-"+thread.ID),
		Class(className),
		Href(fmt.Sprintf("/threads/%s", thread.ID)),

		Span(Class("thread-item-text"), Text(thread.Title)),
		If(thread.IsStreaming,
			Span(Class("thread-item-spinner"), iconSpinner()),
		),
		Button(
			Type("button"),
			Class("archive-btn"),
			Data("on:click", fmt.Sprintf(`
				evt.preventDefault();
				evt.stopPropagation();
				@post('/threads/%s/archive');
			`, thread.ID)),
			iconArchive(),
		),
	)
}
