package views

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"datastar-chat/db"

	"github.com/google/uuid"
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

func (r *Renderer) RenderIndex(w io.Writer, data IndexData) {
	renderNode(w, r.indexPage(data))
}

func (r *Renderer) RenderThread(w io.Writer, data ThreadData) {
	renderNode(w, r.threadPage(data))
}

func renderNode(w io.Writer, node g.Node) {
	err := node.Render(w)
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func RenderThreads(w io.Writer, threads []db.Thread, activeThreadID uuid.UUID) {
	for _, t := range threads {
		renderNode(w, threadListItem(t, t.ID == activeThreadID))
	}
}

func (r *Renderer) indexPage(data IndexData) g.Node {
	return r.page(
		sidebar(data.Threads, uuid.Nil),
		threadContentsEmpty(),
	)
}

func (r *Renderer) threadPage(data ThreadData) g.Node {
	return r.page(
		sidebar(data.Threads, data.Thread.ID),
		threadContents(data.Messages, data.Thread.ID),
	)
}

func (r *Renderer) page(body ...g.Node) g.Node {
	return components.HTML5(components.HTML5Props{
		Title:    "datastar.chat",
		Language: "en",
		Head: []g.Node{
			Link(Rel("preconnect"), Href("https://fonts.googleapis.com")),
			Link(Rel("preconnect"), Href("https://fonts.gstatic.com"), g.Attr("crossorigin")),
			Link(
				Rel("stylesheet"),
				Href("https://fonts.googleapis.com/css2?family=Fira+Code:wght@300..700&display=swap"),
			),
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+r.assetVersion)),
			Script(
				Type("module"),
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.7/bundles/datastar.js"),
			),
		},
		Body: []g.Node{
			// Pre(g.Attr("data-json-signals")),

			Div(Class("app"), g.Group(body)),
		},
	})
}

func sidebar(threads []db.Thread, activeThreadID uuid.UUID) g.Node {
	return Aside(
		Class("sidebar"),
		g.If(activeThreadID != uuid.Nil,
			g.Attr("data-init", fmt.Sprintf("@get('/threads/%s/events')", activeThreadID)),
		),

		Div(Class("sidebar-header"),
			Span(
				Class("logo"),
				g.Text("datastar"),
				Span(Class("logo-accent"), g.Text(".chat")),
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
		ID("thread-"+thread.ID.String()),
		Href("/threads/"+thread.ID.String()),
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
	return threadContents(nil, uuid.Nil)
}

func threadContents(messages []db.Message, threadID uuid.UUID) g.Node {
	var contents g.Node
	if len(messages) == 0 {
		contents = Div(
			Div(
				ID("messages"),
				Class("main-body empty-thread"),
				Div(
					Class("empty-state"),
					H1(Class("empty-state-title"), g.Text("How can I help?")),
					P(Class("empty-state-subtitle"),
						g.Text("Start a new conversation with datastar.chat."),
					),
				),
			),
			newThreadChatInput(),
		)
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
	threadID uuid.UUID
}

func chatInput(props chatInputProps) g.Node {
	var threadIDSignal g.Node
	if props.threadID != uuid.Nil {
		threadIDSignal = g.Attr("data-signals:threadId",
			fmt.Sprintf("'%s'", props.threadID.String()))
	}

	return Div(
		Class("chat-input-area"),
		threadIDSignal,

		Div(
			Class("chat-input-box"),
			Textarea(
				AutoFocus(),
				Name("message"),
				Placeholder("Type your message here..."),
				g.Attr("data-bind:message"),
				g.Attr("data-indicator:loading"),
				g.Attr("data-on:keydown", `
					if (evt.key !== 'Enter') return;
					if (evt.shiftKey) return;

					@post('/chat/send');
				`),
			),
		),
	)
}

func existingThreadChatInput(threadID uuid.UUID) g.Node {
	return chatInput(chatInputProps{
		threadID: threadID,
	})
}

func newThreadChatInput() g.Node {
	return chatInput(chatInputProps{uuid.Nil})
}

func userMessage(messageID uuid.UUID, content string) g.Node {
	return Div(
		ID("msg-"+messageID.String()),
		Class("message user-message"),
		g.Text(content),
	)
}

func RenderAssistantMessage(w io.Writer, messageID uuid.UUID, contentHTML string) {
	renderNode(w, assistantMessage(messageID, contentHTML))
}

func RenderUserMessage(w io.Writer, messageID uuid.UUID, contentHTML string) {
	renderNode(w, userMessage(messageID, contentHTML))
}

func assistantMessage(messageID uuid.UUID, contentHTML string) g.Node {
	return Div(
		ID("msg-"+messageID.String()),
		Class("message assistant-message"),
		g.Raw(contentHTML),
	)
}
