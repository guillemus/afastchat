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

type (
	Thread struct {
		ID          string
		IsStreaming bool
		Title       string
	}
	Message struct {
		Type string
		ID   string
		// markdown rendered content
		MdContent string
		ToolName  string
	}
)

type IndexData struct {
	Threads []Thread
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []Message
	Streaming bool
}

func (x *Renderer) appLayout(body ...Node) Node {
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
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.8/bundles/datastar.js"),
			),
		},
		Body: []Node{
			DebugSignals(),
			Div(Class("app"), Group(body)),
		},
	})
}

func (x *Renderer) RenderIndexPage(w io.Writer, data IndexData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		sidebar(data.Threads, "", doScrollSidebar),
		threadContentsEmpty(),
	))
}

func (x *Renderer) RenderThreadPage(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		sidebar(data.Threads, data.Thread.ID, doScrollSidebar),
		threadContents(data.Messages, data.Thread.ID, doScrollSidebar, data.Streaming),
	))
}

func renderNode(w io.Writer, node Node) {
	err := node.Render(w)
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func PatchThreads(sse *datastar.ServerSentEventGenerator, threads []Thread, activeThreadID string) {
	var sb strings.Builder
	for _, it := range threads {
		renderNode(&sb, threadListItem(it, it.ID == activeThreadID))
	}

	sse.PatchElements(sb.String(), datastar.WithSelectorID("thread-items"))
}

func PatchThreadItem(sse *datastar.ServerSentEventGenerator, thread Thread) {
	var sb strings.Builder
	renderNode(&sb, threadListItem(thread, true))
	sse.PatchElements(sb.String())
}

func PatchThreadMessages(
	sse *datastar.ServerSentEventGenerator, messages []Message, doScrollSidebar bool) {
	var sb strings.Builder
	renderNode(&sb, ThreadMessages(messages, doScrollSidebar))
	sse.PatchElements(sb.String(), datastar.WithSelectorID("messages"))
}
