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

func Sidebar(x SidebarProps) Node {
	return Aside(
		Class("sidebar flex min-h-screen w-[260px] min-w-[260px] max-w-[260px] shrink-0 flex-col overflow-hidden border-r border-neutral-800 bg-stone-950 text-[length:0.875rem]"),
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
	return Div(Class("flex items-center justify-between px-3 pt-4 pb-3"), AppLogo())
}

func AppLogo() Node {
	return Span(
		Class("font-normal tracking-[-0.02em]"),
		Text("afastchat"),
		Span(Class("text-orange-500"), Text(".com")),
	)
}

func NewChatButton() Node {
	return A(Href("/"), Class("mx-3 mb-3 flex items-center justify-center gap-2 rounded-lg border-0 bg-orange-500 px-3 py-2 font-normal text-white no-underline transition hover:brightness-110"), iconPlus("h-3.5 w-3.5 shrink-0"), Text("New Chat"))
}

func ThreadSearch() Node {
	return Div(Class("relative mx-3 mb-2"),
		iconSearch("pointer-events-none absolute left-2 top-1/2 h-[13px] w-[13px] -translate-y-1/2 text-neutral-500"),
		Input(
			Class("w-full rounded-lg border border-transparent bg-transparent py-1 pr-2 pl-[calc(var(--spacing)*2+13px+var(--spacing)*2)] font-[inherit] text-[inherit] text-neutral-400 outline-none transition-colors focus:border-neutral-800 focus:bg-neutral-900 placeholder:text-neutral-500"),
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
		Class("flex items-center gap-2 border-t border-neutral-800 p-3"),
		UserAvatar(email),
		Form(Class("ml-auto"), Method("post"), Action("/logout"), LogoutButton()),
	)
}

func UserAvatar(email string) Node {
	return Div(Class("flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-full bg-orange-500 font-normal text-white"), Text(userInitial(email)))
}

func userInitial(email string) string {
	if email == "" {
		return "?"
	}
	return string([]rune(email)[0])
}

func LogoutButton() Node {
	return Button(Type("submit"), Class("min-h-8 cursor-pointer rounded-lg border border-neutral-800 bg-transparent px-2 font-[inherit] text-[inherit] text-neutral-400 hover:bg-neutral-900 hover:text-neutral-100"), Text("Log out"))
}
