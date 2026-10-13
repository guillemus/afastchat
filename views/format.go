package views

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type threadGroup struct {
	Label   string
	Threads []Thread
}

func groupThreads(threads []Thread, now time.Time) []threadGroup {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	bounds := []struct {
		label string
		since time.Time
	}{
		{"Today", today},
		{"Yesterday", today.AddDate(0, 0, -1)},
		{"Previous 7 days", today.AddDate(0, 0, -7)},
		{"Previous 30 days", today.AddDate(0, 0, -30)},
		{"Older", time.Time{}},
	}

	var groups []threadGroup
	for _, thread := range threads {
		label := bounds[len(bounds)-1].label
		for _, bound := range bounds {
			if !thread.CreatedAt.Before(bound.since) {
				label = bound.label
				break
			}
		}
		if len(groups) == 0 || groups[len(groups)-1].Label != label {
			groups = append(groups, threadGroup{Label: label})
		}
		groups[len(groups)-1].Threads = append(groups[len(groups)-1].Threads, thread)
	}
	return groups
}

func initials(viewer Viewer) string {
	var out []rune
	for word := range strings.FieldsSeq(viewer.Name) {
		out = append(out, []rune(word)[0])
		if len(out) == 2 {
			break
		}
	}
	if len(out) == 0 && viewer.Email != "" {
		out = append(out, []rune(viewer.Email)[0])
	}
	if len(out) == 0 {
		return "?"
	}
	return strings.ToUpper(string(out))
}

func unixMilli(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

func ToolCallDisplay(name, args string) (string, string) {
	var a struct {
		Query string `json:"query"`
		URL   string `json:"url"`
	}
	_ = json.Unmarshal([]byte(args), &a)
	switch name {
	case "web_search":
		return "Searching", a.Query
	case "web_fetch":
		return "Reading", a.URL
	default:
		return "Using " + name, ""
	}
}

func toolCallDone(name, args string) (string, string) {
	_, target := ToolCallDisplay(name, args)
	switch name {
	case "web_search":
		return "Searched", target
	case "web_fetch":
		return "Read", target
	default:
		return "Used " + name, target
	}
}

func workID(props AssistantMessagePartsProps) string {
	if props.Working {
		return "work-" + props.ID + "-live"
	}
	return "work-" + props.ID + "-done"
}

type locationKey struct{}

func WithLocation(ctx context.Context, loc *time.Location) context.Context {
	return context.WithValue(ctx, locationKey{}, loc)
}

func viewerNow(ctx context.Context) time.Time {
	if loc, ok := ctx.Value(locationKey{}).(*time.Location); ok {
		return time.Now().In(loc)
	}
	return time.Now()
}

func clockTime(ctx context.Context, t time.Time) string {
	loc, ok := ctx.Value(locationKey{}).(*time.Location)
	if !ok {
		return ""
	}
	return t.In(loc).Format("15:04")
}

func timeAgo(t, now time.Time) string {
	minutes := int(now.Sub(t) / time.Minute)
	switch {
	case minutes < 1:
		return "now"
	case minutes < 60:
		return strconv.Itoa(minutes) + "m"
	case minutes < 60*24:
		return strconv.Itoa(minutes/60) + "h"
	case minutes < 60*24*7:
		return strconv.Itoa(minutes/(60*24)) + "d"
	case minutes < 60*24*7*52:
		return strconv.Itoa(minutes/(60*24*7)) + "w"
	default:
		return strconv.Itoa(minutes/(60*24*365)) + "y"
	}
}

func elapsedLabel(since, now time.Time) string {
	return strconv.Itoa(max(1, int(now.Sub(since).Round(time.Second)/time.Second))) + "s"
}
