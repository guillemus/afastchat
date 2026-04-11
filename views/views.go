package views

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"afastchat.com/db"

	g "maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
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

type IndexData struct {
	Threads []db.Thread
}

type ThreadData struct {
	Threads   []db.Thread
	Thread    db.Thread
	Messages  []db.Message
	Streaming bool
}

func (x *Renderer) RenderIndex(w io.Writer, data IndexData) {
	renderNode(w, x.indexPage(data))
}

func (x *Renderer) RenderThread(w io.Writer, data ThreadData) {
	renderNode(w, x.threadPage(data))
}

func renderNode(w io.Writer, node g.Node) {
	err := node.Render(w)
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func RenderThreads(w io.Writer, threads []db.Thread, activeThreadID string) {
	for _, it := range threads {
		renderNode(w, threadListItem(it, it.ID == activeThreadID))
	}
}

func (x *Renderer) indexPage(data IndexData) g.Node {
	return x.page(
		sidebar(data.Threads, ""),
		threadContentsEmpty(),
	)
}

func (x *Renderer) threadPage(data ThreadData) g.Node {
	return x.page(
		sidebar(data.Threads, data.Thread.ID),
		threadContents(data.Messages, data.Thread.ID),
	)
}

func (x *Renderer) page(body ...g.Node) g.Node {
	return components.HTML5(components.HTML5Props{
		Title:    "afastchat",
		Language: "en",
		Head: []g.Node{
			Link(Rel("preconnect"), Href("https://fonts.googleapis.com")),
			Link(Rel("preconnect"), Href("https://fonts.gstatic.com"), g.Attr("crossorigin")),
			Link(
				Rel("stylesheet"),
				Href("https://fonts.googleapis.com/css2?family=Fira+Code:wght@300..700&display=swap"),
			),
			Link(Rel("stylesheet"), Href("https://unpkg.com/open-props@1.7.23/open-props.min.css")),
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+x.assetVersion)),
			Script(
				Type("module"),
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.7/bundles/datastar.js"),
			),
			Script(Type("module"), Src("/public/main.js?v="+x.assetVersion)),
		},
		Body: []g.Node{
			// DebugSignals(),
			Div(Class("app"), g.Group(body)),
		},
	})
}

func sidebar(threads []db.Thread, activeThreadID string) g.Node {
	return Aside(
		Class("sidebar"),
		g.If(activeThreadID != "",
			g.Attr("data-init", fmt.Sprintf("@get('/threads/%s/events')", activeThreadID)),
		),

		Div(Class("sidebar-header"),
			Span(
				Class("logo"),
				g.Text("afastchat"),
				Span(Class("logo-accent"), g.Text(".com")),
			),
		),
		A(Href("/"), Class("new-chat-btn"),
			iconPlus(),
			g.Text("New Chat"),
		),
		Div(Class("search-box"),
			iconSearch(),
			Input(
				Type("text"),
				Name("thread-search"),
				Placeholder("Search threads..."),
				g.Attr("data-bind:threadSearch"),
				g.Attr("data-on:input__debounce.150ms", "@get('/threads/search')"),
			),
		),
		Div(Class("thread-list"),
			Div(
				ID("thread-section-recent"),
				Class("thread-section-label"),
				g.Text("Recent"),
			),
			Div(ID("thread-items"),
				g.Map(threads, func(thread db.Thread) g.Node {
					return threadListItem(thread, thread.ID == activeThreadID)
				}),
			),
		),
		Div(
			Class("sidebar-footer"),
			Div(Class("avatar"), g.Text("G")),
			Span(Class("sidebar-footer-name"), g.Text("Guillem")),
		),
	)
}

func threadListItem(thread db.Thread, isActive bool) g.Node {
	className := "thread-item"
	if isActive {
		className += " active"
	}

	return A(
		ID("thread-"+thread.ID),
		Href("/threads/"+thread.ID),
		Class(className),
		Span(Class("thread-item-text"), g.Text(thread.Title)),
		Button(
			Type("button"),
			Class("archive-btn"),
			g.Attr("data-on:click", fmt.Sprintf(`
				evt.preventDefault();
				evt.stopPropagation();
				@post('/threads/%s/archive');
			`, thread.ID)),
			iconArchive(),
		),
	)
}

func threadContentsEmpty() g.Node {
	return threadContents(nil, "")
}

func threadContents(messages []db.Message, threadID string) g.Node {
	var contents g.Node
	if len(messages) == 0 {
		contents = g.Group{
			Div(
				ID("messages"),
				Class("main-body empty-thread"),
				Div(
					Class("empty-state"),
					H1(Class("empty-state-title"), g.Text("How can I help?")),
					P(Class("empty-state-subtitle"),
						g.Text("Start a new conversation with afastchat."),
					),
				),
			),
			newThreadChatInput(),
		}
	} else {
		contents = g.Group{
			Div(
				ID("messages"),
				Class("main-body"),
				g.Map(messages, func(message db.Message) g.Node {
					switch message.Role {
					case "user":
						return userMessage(message.ID, message.Content)
					case "assistant":
						return assistantMessage(message.ID, message.Content)
					default:
						return nil
					}
				}),
			),
			existingThreadChatInput(threadID),
		}
	}

	return Main(Class("main"), contents)
}

type chatInputProps struct {
	threadID string
}

func chatInput(props chatInputProps) g.Node {
	return Div(
		Class("chat-input"),
		g.If(props.threadID != "",
			g.Attr("data-signals:threadId", fmt.Sprintf("'%s'", props.threadID)),
		),
		g.Attr("data-signals:is-streaming", "false"),
		g.Attr("data-signals:message", "''"),
		g.Attr("data-computed:can-submit", "$message.trim() !== ''"),
		Div(
			Class("chat-input-box"),
			Textarea(
				AutoFocus(),
				Name("message"),
				Placeholder("Type your message here..."),
				g.Attr("data-bind:message"),
				g.Attr("data-on:keydown", `
					if (evt.key !== 'Enter' || evt.shiftKey) return;
					if (!$canSubmit) return;

					@post('/chat/send');
				`),
			),
			Div(Class("chat-input-actions"),
				Button(
					g.Attr("data-show", "!$isStreaming"),
					Disabled(),
					Class("chat-input-submit-btn"),
					Type("submit"),
					g.Attr("data-attr:disabled", "!$canSubmit || $isStreaming"),
					g.Attr("data-on:click", `
						@post('/chat/send')
					`),
					iconArrowUp(),
				),
				Button(
					g.Attr("data-show", "$isStreaming"),
					Class("chat-input-submit-btn"),
					Type("submit"),
					g.Attr("data-on:click", `
						@post('/chat/stop')
					`),
					iconSquare(),
				),
			),
		),
	)
}

func existingThreadChatInput(threadID string) g.Node {
	return chatInput(chatInputProps{
		threadID: threadID,
	})
}

func newThreadChatInput() g.Node {
	return chatInput(chatInputProps{""})
}

func userMessage(messageID string, content string) g.Node {
	return Div(
		ID("msg-"+messageID),
		Class("message user-message"),
		g.Text(content),
	)
}

func RenderAssistantMessage(w io.Writer, messageID string, contentHTML string) {
	renderNode(w, assistantMessage(messageID, contentHTML))
}

func RenderUserMessage(w io.Writer, messageID string, contentHTML string) {
	renderNode(w, userMessage(messageID, contentHTML))
}

func assistantMessage(messageID string, contentHTML string) g.Node {
	return Div(
		ID("msg-"+messageID),
		Class("message assistant-message"),
		g.Raw(contentHTML),
	)
}

func DebugSignals() g.Node {
	return Div(
		Style(`
			background: black;
			position: absolute;	
			top: 0;
			right: 0;
		`),
		Pre(g.Attr("data-json-signals")),
	)
}
