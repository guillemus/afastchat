package views

import (
	stdhtml "html"
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.conversation-minimap {
		width: 100%;
		display: flex;
		flex-direction: column;
		min-height: 100vh;
		background: color-mix(in oklch, var(--bg-sidebar) 90%, black);
		border-left: 1px solid var(--border-color);
		overflow: hidden;
	}
	.minimap-header {
		display: flex;
		align-items: center;
		gap: var(--size-2);
		padding: var(--size-4) var(--size-3);
		font-weight: var(--font-weight-6);
		letter-spacing: -0.02em;
	}
	.minimap-icon {
		color: var(--text-secondary);
		font-size: var(--font-size-2);
	}
	.minimap-list {
		flex: 1;
		overflow-y: auto;
		padding: var(--size-2) var(--size-3) var(--size-6);
		scrollbar-width: thin;
		scrollbar-color: var(--gray-9) transparent;
	}
	.minimap-track {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--size-3);
	}
	.minimap-empty {
		display: grid;
		place-items: center;
		min-height: 40vh;
		color: var(--text-secondary);
		text-align: center;
		line-height: var(--font-lineheight-3);
	}
	.minimap-node {
		position: relative;
		width: min(100%, 340px);
	}
	.minimap-node:not(:last-child)::after {
		content: "";
		position: absolute;
		left: 50%;
		top: 100%;
		width: 1px;
		height: var(--size-3);
		background: var(--gray-8);
	}
	.minimap-card {
		display: flex;
		flex-direction: column;
		gap: var(--size-1);
		width: 100%;
		min-height: 0;
		padding: var(--size-2) var(--size-3);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-2);
		background: color-mix(in oklch, var(--gray-11) 70%, transparent);
		color: var(--text-secondary);
		font: inherit;
		font-size: var(--font-sidebar-size);
		text-align: left;
		cursor: pointer;
		box-shadow: var(--shadow-1);
		transition:
			background 0.12s var(--ease-3),
			border-color 0.12s var(--ease-3),
			box-shadow 0.12s var(--ease-3),
			color 0.12s var(--ease-3);
	}
	.minimap-card:hover,
	.minimap-card:focus-visible {
		outline: none;
		border-color: color-mix(in oklch, var(--bg-button) 35%, var(--border-color));
		background: color-mix(in oklch, var(--gray-10) 62%, transparent);
		color: var(--text-primary);
	}
	.minimap-card.is-visible {
		border-color: color-mix(in oklch, var(--bg-button) 30%, var(--border-color));
		background: color-mix(in oklch, var(--bg-button) 7%, var(--gray-11));
		color: var(--text-primary);
		box-shadow: var(--shadow-1);
	}
	.minimap-card-top {
		display: flex;
		align-items: center;
		gap: var(--size-2);
		flex-shrink: 0;
	}
	.minimap-dot {
		width: 11px;
		height: 11px;
		border-radius: var(--radius-round);
		background: var(--bg-button);
		flex-shrink: 0;
	}
	.minimap-card.assistant .minimap-dot {
		background: var(--green-6);
	}
	.minimap-role {
		font-weight: var(--font-weight-6);
		color: var(--text-primary);
	}
	.minimap-preview {
		display: -webkit-box;
		overflow: hidden;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		line-height: var(--font-lineheight-1);
	}
	@media (max-width: 1180px) {
		.app {
			grid-template-columns: 260px minmax(0, 1fr);
		}
		.conversation-minimap {
			display: none;
		}
	}
`)

func MessageMinimap(messages []MessageGroup) Node {
	items := minimapItems(messages)

	return Aside(
		ID("conversation-minimap"),
		Class("conversation-minimap"),
		Data("init", "updateVisibleMinimapMessages()"),
		Div(
			Class("minimap-header"),
			Span(Class("minimap-icon"), Text("⌘")),
			Text("Conversation Tree"),
		),
		Nav(
			Class("minimap-list"),
			Attr("aria-label", "Conversation minimap"),
			If(len(items) == 0,
				Div(Class("minimap-empty"), Text("Start chatting to grow the tree.")),
			),
			If(len(items) > 0,
				Div(
					Class("minimap-track"),
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
		Class("minimap-node"),
		Button(
			Type("button"),
			Classes{
				"minimap-card": true,
				"assistant":    item.Role == "Assistant",
			},
			Attr("aria-label", "Scroll to "+item.Role+" message"),
			Attr("data-message-id", item.ID),
			Data("on:click", `
				const messages = document.getElementById('messages');
				const target = document.getElementById('msg-`+item.ID+`');
				messages.scrollTop += target.getBoundingClientRect().top - messages.getBoundingClientRect().top;
				updateVisibleMinimapMessages();
			`),
			Div(
				Class("minimap-card-top"),
				Span(Class("minimap-dot")),
				Span(Class("minimap-role"), Text(item.Role)),
			),
			Span(Class("minimap-preview"), Text(item.Preview)),
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
