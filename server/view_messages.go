package server

import (
	"encoding/json"
	"html"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"afastchat.com/db"
	"afastchat.com/views"
)

func mapViewMessages(messages []db.Message, streaming, largeHeight bool) []views.MessageGroup {
	out := make([]views.MessageGroup, 0, len(messages))
	var turn []db.Message
	var startedAt time.Time

	flush := func(isLast bool) {
		if len(turn) == 0 {
			return
		}
		out = append(out, mapAssistantTurn(turn, startedAt, streaming && isLast))
		turn = nil
	}

	for _, it := range messages {
		switch it.Type {
		case db.SystemMessage, db.ToolOutput:
		case db.UserMessage:
			flush(false)
			startedAt = it.CreatedAt
			out = append(out, views.UserMessageProps{
				ID:        it.ID,
				MdContent: renderMarkdownCached(it.Content),
				CreatedAt: it.CreatedAt,
			})
		case db.AssistantMessage, db.AssistantReasoning, db.AssistantToolCall, db.AssistantSources:
			turn = append(turn, it)
		}
	}
	flush(true)

	if largeHeight {
		for i := len(out) - 1; i >= 0; i-- {
			if msg, ok := out[i].(views.AssistantMessagePartsProps); ok {
				msg.LargeHeight = true
				out[i] = msg
				break
			}
		}
	}

	return out
}

func mapAssistantTurn(turn []db.Message, startedAt time.Time, streaming bool) views.AssistantMessagePartsProps {
	render := renderMarkdownCached
	if streaming {
		render = renderStreamingMarkdown
	}

	var refs []db.SourceRef
	var sourcesID string
	var sourcesUpdatedAt time.Time
	for _, it := range turn {
		if it.Type != db.AssistantSources {
			continue
		}
		sourcesID = it.ID
		sourcesUpdatedAt = it.UpdatedAt
		var parsed []db.SourceRef
		if err := json.Unmarshal([]byte(it.Content), &parsed); err != nil {
			slog.Error("decode sources", "message_id", it.ID, "err", err)
			continue
		}
		refs = append(refs, parsed...)
	}

	lastWork := -1
	for i, it := range turn {
		switch it.Type {
		case db.AssistantReasoning, db.AssistantToolCall, db.AssistantSources:
			lastWork = i
		case db.AssistantMessage:
			if it.CreatedAt.Before(sourcesUpdatedAt) {
				lastWork = i
			}
		case db.UserMessage, db.SystemMessage, db.ToolOutput:
		}
	}

	var answer strings.Builder
	var answerStartedAt time.Time
	for _, it := range turn[lastWork+1:] {
		if answerStartedAt.IsZero() && strings.TrimSpace(it.Content) != "" {
			answerStartedAt = it.CreatedAt
		}
		answer.WriteString(it.Content)
	}
	answerText := strings.TrimSpace(answer.String())

	props := views.AssistantMessagePartsProps{
		ID:        turn[0].ID,
		StartedAt: startedAt,
		Working:   streaming && answerText == "",
		Work:      mapWorkParts(turn[:lastWork+1], render),
	}

	for _, ref := range refs {
		props.Work = append(props.Work, views.SourceReadPartProps{
			ID:     sourcesID + "-" + ref.URL,
			Title:  sourceTitle(ref),
			Domain: sourceDomain(ref.URL),
			URL:    ref.URL,
		})
	}

	if answerText != "" {
		props.AnswerHTML, props.Sources = applyCitations(render(answerText), refs)
	}

	switch {
	case !answerStartedAt.IsZero():
		props.WorkSeconds = elapsedSeconds(startedAt, answerStartedAt)
	case !streaming:
		props.WorkSeconds = elapsedSeconds(startedAt, turn[len(turn)-1].UpdatedAt)
	}

	if props.Working {
		props.Status = workStatus(turn, refs)
	}

	return props
}

func mapWorkParts(messages []db.Message, render func(string) string) []views.AssistantPart {
	var parts []views.AssistantPart
	var narration strings.Builder
	var narrationID string

	flushNarration := func() {
		if text := strings.TrimSpace(narration.String()); text != "" {
			parts = append(parts, views.AssistantTextPartProps{ID: narrationID, MdContent: render(text)})
		}
		narration.Reset()
		narrationID = ""
	}

	for _, it := range messages {
		switch it.Type {
		case db.AssistantMessage:
			if narrationID == "" {
				narrationID = it.ID
			}
			narration.WriteString(it.Content)
		case db.AssistantReasoning:
			flushNarration()
			if strings.TrimSpace(it.Content) == "" {
				continue
			}
			parts = append(parts, views.ReasoningPartProps{
				ID:        it.ID,
				MdContent: render(it.Content),
				Seconds:   elapsedSeconds(it.CreatedAt, it.UpdatedAt),
			})
		case db.AssistantToolCall:
			flushNarration()
			parts = append(parts, views.ToolCallPartProps{ID: it.ID, Name: it.ToolName, Args: it.Content})
		case db.UserMessage, db.SystemMessage, db.ToolOutput, db.AssistantSources:
		}
	}
	flushNarration()

	return parts
}

func workStatus(turn []db.Message, refs []db.SourceRef) views.Status {
	if len(turn) == 0 {
		return views.Status{Key: "start", Verb: "Getting started"}
	}

	last := turn[len(turn)-1]
	switch last.Type {
	case db.AssistantReasoning:
		return views.Status{
			Key: last.ID, Icon: views.StatusThinking, Verb: "Thinking", Words: tailWords(last.Content, 160),
		}
	case db.AssistantToolCall:
		name, target := views.ToolCallDisplay(last.ToolName, last.Content)
		icon := views.StatusSearching
		if last.ToolName == "web_fetch" {
			icon = views.StatusReading
		}
		return views.Status{Key: last.ID, Icon: icon, Verb: name, Target: target}
	case db.AssistantSources:
		if len(refs) > 0 {
			ref := refs[len(refs)-1]
			return views.Status{
				Key: last.ID + strconv.Itoa(len(refs)), Icon: views.StatusReading, Verb: "Reading", Target: sourceTitle(ref),
			}
		}
	case db.AssistantMessage:
		if strings.TrimSpace(last.Content) == "" {
			break
		}
		return views.Status{
			Key: last.ID, Icon: views.StatusThinking, Verb: "Working", Words: tailWords(last.Content, 160),
		}
	case db.UserMessage, db.SystemMessage, db.ToolOutput:
	}

	return views.Status{Key: "start", Verb: "Getting started"}
}

func tailWords(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[len(runes)-limit:])
}

func elapsedSeconds(from, to time.Time) int {
	if from.IsZero() || to.IsZero() {
		return 1
	}
	return max(1, int(to.Sub(from).Round(time.Second).Seconds()))
}

func sourceDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return strings.TrimPrefix(parsed.Hostname(), "www.")
}

func sourceTitle(ref db.SourceRef) string {
	if title := strings.TrimSpace(ref.Title); title != "" {
		return title
	}
	return sourceDomain(ref.URL)
}

var (
	anchorRe = regexp.MustCompile(`<a href="([^"]*)">(.*?)</a>`)
	tagRe    = regexp.MustCompile(`<[^>]*>`)
)

func normalizeSourceURL(rawURL string) string {
	return strings.TrimSuffix(strings.SplitN(rawURL, "#", 2)[0], "/")
}

func applyCitations(answerHTML string, refs []db.SourceRef) (string, []views.Source) {
	byURL := make(map[string]db.SourceRef, len(refs))
	for _, ref := range refs {
		byURL[normalizeSourceURL(ref.URL)] = ref
	}

	numbers := map[string]int{}
	var sources []views.Source
	addSource := func(key string, ref db.SourceRef) int {
		if n, ok := numbers[key]; ok {
			return n
		}
		n := len(sources) + 1
		numbers[key] = n
		sources = append(sources, views.Source{
			Number: n, URL: ref.URL, Title: sourceTitle(ref), Domain: sourceDomain(ref.URL),
		})
		return n
	}

	out := anchorRe.ReplaceAllStringFunc(answerHTML, func(match string) string {
		groups := anchorRe.FindStringSubmatch(match)
		href := html.UnescapeString(groups[1])
		key := normalizeSourceURL(href)
		label := strings.TrimSpace(html.UnescapeString(tagRe.ReplaceAllString(groups[2], "")))
		link := strings.Replace(match, "<a ", `<a target="_blank" rel="noopener noreferrer" `, 1)

		ref, known := byURL[key]
		bareLabel := !strings.ContainsAny(label, " \t\n")
		if !known {
			if len(refs) == 0 || !bareLabel || !strings.HasPrefix(href, "http") {
				return link
			}
			ref = db.SourceRef{URL: href}
		}

		n := addSource(key, ref)
		chip := `<a class="cite" target="_blank" rel="noopener noreferrer" href="` + groups[1] +
			`" title="` + html.EscapeString(sourceTitle(ref)) + `">` + strconv.Itoa(n) + `</a>`
		if bareLabel {
			return chip
		}
		return groups[2] + chip
	})

	for _, ref := range refs {
		addSource(normalizeSourceURL(ref.URL), ref)
	}

	return out, sources
}
