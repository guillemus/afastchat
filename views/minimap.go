package views

import (
	stdhtml "html"
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func MessageMinimap(messages []MessageGroup) Node {
	items := minimapItems(messages)

	return Aside(
		ID("conversation-minimap"),
		Class("conversation-minimap flex min-h-screen w-full flex-col overflow-hidden border-l border-[var(--border-color)] bg-[color-mix(in_oklch,var(--bg-sidebar)_90%,black)] max-[1180px]:hidden"),
		Div(
			Class("flex items-center gap-2 px-3 py-4 font-semibold tracking-[-0.02em]"),
			Span(Class("text-base text-[var(--text-secondary)]"), Text("⌘")),
			Text("Conversation Tree"),
		),
		Nav(
			Class("minimap-list flex-1 overflow-y-auto px-3 pt-2 pb-6 [scrollbar-color:var(--gray-9)_transparent] [scrollbar-width:thin]"),
			Attr("aria-label", "Conversation minimap"),
			If(len(items) == 0,
				Div(Class("grid min-h-[40vh] place-items-center text-center leading-7 text-[var(--text-secondary)]"), Text("Start chatting to grow the tree.")),
			),
			If(len(items) > 0,
				Div(
					Class("flex flex-col items-center gap-3"),
					Map(items, minimapItemNode),
				),
			),
		),
	)
}

type minimapItem struct {
	ID      string
	Role    string
	Preview string
}

func minimapItems(messages []MessageGroup) []minimapItem {
	items := make([]minimapItem, 0, len(messages))
	for _, message := range messages {
		switch it := message.(type) {
		case UserMessageProps:
			items = append(items, minimapItem{ID: it.ID, Role: "User", Preview: minimapPreview(it.MdContent)})
		case AssistantMessagePartsProps:
			preview := minimapPreview(assistantPreview(it.Parts))
			if preview != "" {
				items = append(items, minimapItem{ID: it.ID, Role: "Assistant", Preview: preview})
			}
		}
	}
	return items
}

func minimapItemNode(item minimapItem) Node {
	return Div(
		Class("minimap-node relative w-[min(100%,340px)]"),
		Button(
			Type("button"),
			Classes{
				"minimap-card flex min-h-0 w-full cursor-pointer flex-col gap-1 rounded-lg border border-[var(--border-color)] bg-[color-mix(in_oklch,var(--gray-11)_70%,transparent)] px-3 py-2 text-left text-[length:var(--font-sidebar-size)] text-[var(--text-secondary)] shadow-[var(--shadow-1)] transition hover:border-[color-mix(in_oklch,var(--bg-button)_35%,var(--border-color))] hover:bg-[color-mix(in_oklch,var(--gray-10)_62%,transparent)] hover:text-[var(--text-primary)] focus-visible:border-[color-mix(in_oklch,var(--bg-button)_35%,var(--border-color))] focus-visible:bg-[color-mix(in_oklch,var(--gray-10)_62%,transparent)] focus-visible:text-[var(--text-primary)] focus-visible:outline-none": true,
			},
			Attr("aria-label", "Scroll to "+item.Role+" message"),
			Data("on:click", `
				const messages = document.getElementById('messages');
				const target = document.getElementById('msg-`+item.ID+`');
				messages.scrollTop += target.getBoundingClientRect().top - messages.getBoundingClientRect().top;
			`),
			Div(
				Class("flex shrink-0 items-center gap-2"),
				Span(Classes{
					"minimap-dot h-[11px] w-[11px] shrink-0 rounded-full bg-[var(--bg-button)]": true,
					"bg-[var(--green-6)]": item.Role == "Assistant",
				}),
				Span(Class("font-semibold text-[var(--text-primary)]"), Text(item.Role)),
			),
			Span(Class("line-clamp-2 overflow-hidden leading-tight"), Text(item.Preview)),
		),
	)
}

func assistantPreview(parts []AssistantPart) string {
	for i := len(parts) - 1; i >= 0; i-- {
		if it, ok := parts[i].(AssistantTextPartProps); ok {
			return it.MdContent
		}
	}
	return ""
}

func minimapPreview(html string) string {
	// TODO: This isn't great, as we could probably build a "message preview" string instead of reparsing the generated html in another layer.
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(stdhtml.UnescapeString(b.String())), " ")
}
