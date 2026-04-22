package views

import (
	"io"
	"log/slog"
	"net/url"
	"strings"

	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Renderer struct {
	assetVersion string
}

func NewRenderer(assetVersion string) *Renderer {
	return &Renderer{
		assetVersion: url.QueryEscape(assetVersion),
	}
}

type Thread struct {
	ID          string
	IsStreaming bool
	Title       string
}

type IndexData struct {
	Threads []Thread
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []MessageGroup
	Streaming bool
}

func (x *Renderer) layout(bodyClass string, body ...Node) Node {
	return HTML5(HTML5Props{
		Title:       "afastchat",
		Description: "",
		Language:    "en",
		Head: []Node{
			Link(Rel("preconnect"), Href("https://fonts.googleapis.com")),
			Link(Rel("preconnect"), Href("https://fonts.gstatic.com"), Attr("crossorigin")),
			Link(
				Rel("stylesheet"),
				Href("https://fonts.googleapis.com/css2?family=Fira+Code:wght@300..700&display=swap"),
			),
			Link(Rel("stylesheet"), Href("https://unpkg.com/open-props@1.7.23/open-props.min.css")),
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+x.assetVersion)),
			Script(Type("module"), Src("/public/main.js?v="+x.assetVersion)),
			Script(
				Type("module"),
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.1/bundles/datastar.js"),
			),
		},
		Body: []Node{
			DebugSignals(),
			Div(Class(bodyClass), Group(body)),
		},
	})
}

func (x *Renderer) appLayout(body ...Node) Node {
	return x.layout("app", body...)
}

func (x *Renderer) guestLayout(body ...Node) Node {
	return x.layout("guest-app", body...)
}

func (x *Renderer) RenderIndexPage(w io.Writer, data IndexData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		Sidebar{data.Threads, "", doScrollSidebar}.Node(),
		NewEmptyMessageThread(),
	))
}

func (x *Renderer) RenderThreadPage(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		Sidebar{data.Threads, data.Thread.ID, doScrollSidebar}.Node(),
		MessageThread{
			Messages:        data.Messages,
			ThreadID:        data.Thread.ID,
			DoScrollSidebar: doScrollSidebar,
			IsStreaming:     data.Streaming,
		}.Node(),
	))
}

func (x *Renderer) RenderLandingPage(w io.Writer) {
	renderNode(w, x.guestLayout(LandingPage()))
}

func (x *Renderer) RenderLoginPage(w io.Writer) {
	renderNode(w, x.guestLayout(LoginPage()))
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
		slog.Error("sse.PatchElements error", "err", err)
		return
	}
}

func PatchThreads(sse *datastar.ServerSentEventGenerator, threads []Thread, activeThreadID string) {
	ssePatch(sse, ThreadList(threads, activeThreadID))
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread) {
	ssePatch(sse, threadListItem(thread, true))
}
