package views

import (
	"io"
	"net/url"

	"afastchat.com/env"
	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Renderer struct {
	env          env.Env
	assetVersion string
}

func NewRenderer(env env.Env) *Renderer {
	return &Renderer{
		env:          env,
		assetVersion: url.QueryEscape(env.AssetVersion),
	}
}

type Thread struct {
	ID          string
	IsStreaming bool
	Title       string
}

type IndexData struct {
	Threads   []Thread
	UserEmail string
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []MessageGroup
	Streaming bool
	UserEmail string
}

var DocumentHead = Static(Group{
	Link(Rel("preload"), Href("/public/fonts/figtree-latin-var.woff2"), As("font"), Type("font/woff2"), Attr("crossorigin")),
	Link(Rel("preload"), Href("/public/fonts/jetbrains-mono-latin-var.woff2"), As("font"), Type("font/woff2"), Attr("crossorigin")),
	Link(Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css")),
	Script(Type("module"), Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.1/bundles/datastar.js")),
	Script(Type("speculationrules"), Raw(`{
		"prefetch": [{
			"where": { "href_matches": "/*" },
			"eagerness": "moderate"
		}]
	}`)),
	// We do it inline so that client functions are available ASAP. Useful for scroll position restoration while rendering.
	JS,
})

func (x *Renderer) layout(body Node) Node {
	return HTML5(HTML5Props{
		// Title:       "afastchat",
		Language: "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+x.assetVersion)),
			DocumentHead,
		},
		Body: []Node{
			DebugSignals(x.env),
			body,
			ToastRegion,
		},
	})
}

func (x *Renderer) appLayout(body ...Node) Node {
	return x.layout(Div(Class("grid h-dvh w-full grid-cols-[260px_minmax(0,1fr)] overflow-hidden"),
		Group(body),
	))
}

func (x *Renderer) guestLayout(body ...Node) Node {
	return x.layout(Div(Class("min-h-dvh"),
		Group(body),
	))
}

func (x *Renderer) RenderIndexPage(w io.Writer, data IndexData) {
	renderNode(w, x.appLayout(
		Sidebar(SidebarProps{data.Threads, "", data.UserEmail}),
		NewEmptyMessageThread(),
	))
}

func (x *Renderer) RenderThreadPage(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		Sidebar(SidebarProps{data.Threads, data.Thread.ID, data.UserEmail}),
		MessageThread(MessageThreadProps{
			Messages:        data.Messages,
			ThreadID:        data.Thread.ID,
			DoScrollSidebar: doScrollSidebar,
			IsStreaming:     data.Streaming,
		}),
	))
}

func (x *Renderer) RenderSettingsPage(w io.Writer, data SettingsData) {
	renderNode(w, x.appLayout(
		Sidebar(SidebarProps{data.Threads, "", data.UserEmail}),
		SettingsPage(data),
	))
}

func (x *Renderer) RenderCheckoutThankYouPage(w io.Writer, data CheckoutThankYouData) {
	renderNode(w, x.guestLayout(CheckoutThankYouPage(data)))
}

func (x *Renderer) RenderLandingPage(w io.Writer) {
	renderNode(w, x.guestLayout(LandingPage))
}

func (x *Renderer) RenderLoginPage(w io.Writer, data LoginPageData) {
	renderNode(w, x.guestLayout(LoginPage(data)))
}

func (x *Renderer) RenderSignupPage(w io.Writer, data SignupPageData) {
	renderNode(w, x.guestLayout(SignupPage(data)))
}

func PatchThreads(sse *datastar.ServerSentEventGenerator, threads []Thread, activeThreadID string) {
	ssePatch(sse, ThreadItems(threads, activeThreadID))
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread) {
	ssePatch(sse, ThreadListItem(thread, true))
}
