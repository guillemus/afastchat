package views

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

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

func (x *Renderer) layout(body Node) Node {
	return HTML5(HTML5Props{
		// Title:       "afastchat",
		Description: "",
		Language:    "en",
		Head: []Node{
			Link(Rel("preload"), Href("/public/fonts/figtree-latin-var.woff2"), As("font"), Type("font/woff2"), Attr("crossorigin")),
			Link(Rel("preload"), Href("/public/fonts/jetbrains-mono-latin-var.woff2"), As("font"), Type("font/woff2"), Attr("crossorigin")),
			Link(Rel("stylesheet"), Href("https://unpkg.com/open-props@1.7.23/open-props.min.css")),
			Link(Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css")),
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+x.assetVersion)),
			Raw(`<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.1/bundles/datastar.js"></script>`),
			Script(Type("speculationrules"), Raw(`{
				"prefetch": [{
					"where": { "href_matches": "/*" },
					"eagerness": "moderate"
				}]
			}`)),
			styles.Node(),
			// We do it inline so that the functions are available ASAP.
			JS,
		},
		Body: []Node{
			DebugSignals(x.env),
			body,
			ToastRegion(),
		},
	})
}

var _ = styles.Style(`
	.app {
		display: grid;
		grid-template-columns: 260px minmax(0, 1fr) 420px;
		width: 100%;
		height: 100vh;
		height: 100dvh;
		overflow: hidden;
	}
	.guest-app {
		min-height: 100vh;
		min-height: 100dvh;
	}
`)

func (x *Renderer) appLayout(body ...Node) Node {
	return x.layout(Div(Class("app"),
		Group(body),
	))
}

func (x *Renderer) guestLayout(body ...Node) Node {
	return x.layout(Div(Class("guest-app"),
		Group(body),
	))
}

func (x *Renderer) RenderIndexPage(w io.Writer, data IndexData) {
	renderNode(w, x.appLayout(
		Sidebar{data.Threads, "", data.UserEmail}.Node(),
		NewEmptyMessageThread(),
	))
}

func (x *Renderer) RenderThreadPage(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		Sidebar{data.Threads, data.Thread.ID, data.UserEmail}.Node(),
		MessageThread{
			Messages:        data.Messages,
			ThreadID:        data.Thread.ID,
			DoScrollSidebar: doScrollSidebar,
			IsStreaming:     data.Streaming,
		}.Node(),
		MessageMinimap(data.Messages),
	))
}

func (x *Renderer) RenderSettingsPage(w io.Writer, data SettingsData) {
	renderNode(w, x.appLayout(
		Sidebar{data.Threads, "", data.UserEmail}.Node(),
		SettingsPage(data),
	))
}

func (x *Renderer) RenderCheckoutThankYouPage(w io.Writer, data CheckoutThankYouData) {
	renderNode(w, x.guestLayout(CheckoutThankYouPage(data)))
}

func (x *Renderer) RenderLandingPage(w io.Writer) {
	renderNode(w, x.guestLayout(LandingPage()))
}

func (x *Renderer) RenderLoginPage(w io.Writer, data LoginPageData) {
	renderNode(w, x.guestLayout(LoginPage(data)))
}

func (x *Renderer) RenderSignupPage(w io.Writer, data SignupPageData) {
	renderNode(w, x.guestLayout(SignupPage(data)))
}

func renderNode(w io.Writer, node Node) {
	if err := node.Render(w); err != nil {
		slog.Error("view render error", "err", err)
	}
}

func ssePatch(sse *datastar.ServerSentEventGenerator, node Node) {
	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		slog.Error("view render error", "err", err)
		return
	}

	if err := sse.PatchElements(sb.String()); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		slog.Error("sse.PatchElements error", "err", err)
		return
	}
}

func sseExecJS(sse *datastar.ServerSentEventGenerator, script string) {
	// Add block scope so that declared variables are not on global scope.
	s := fmt.Sprintf("{ %s }", script)
	if err := sse.ExecuteScript(s); err != nil {
		slog.Error("sse.ExecuteScript", "err", err)
	}
}

func PatchThreads(sse *datastar.ServerSentEventGenerator, threads []Thread, activeThreadID string) {
	ssePatch(sse, ThreadItems(threads, activeThreadID))
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread) {
	ssePatch(sse, ThreadListItem(thread, true))
}
