package views

import (
	"context"
	"strings"
	"time"
)

const (
	SidebarHeaderHeight  = 40
	SidebarRowHeight     = 31
	SidebarBuffer        = 1000
	SidebarDefaultHeight = 1000
	SidebarMaxHeight     = 20000
	sidebarMaxY          = 100_000_000
)

var threadGroupLabels = [...]string{"Today", "Yesterday", "Previous 7 days", "Previous 30 days", "Older"}

func ThreadGroupBounds(ctx context.Context) []time.Time {
	now := viewerNow(ctx)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return []time.Time{
		today,
		today.AddDate(0, 0, -1),
		today.AddDate(0, 0, -7),
		today.AddDate(0, 0, -30),
	}
}

type SidebarPlan struct {
	Items  []SidebarPlanItem
	Top    int
	Bottom int
	Total  int
}

type SidebarPlanItem struct {
	Group int
	Row   int
}

func clampSidebarView(y, h int) (int, int) {
	return min(max(y, -sidebarMaxY), sidebarMaxY), min(max(h, 0), SidebarMaxHeight)
}

func PlanSidebar(counts []int, y, h int) SidebarPlan {
	y, h = clampSidebarView(y, h)
	var plan SidebarPlan
	for _, count := range counts {
		if count > 0 {
			plan.Total += SidebarHeaderHeight + count*SidebarRowHeight
		}
	}

	lo := max(0, y-SidebarBuffer)
	hi := min(plan.Total, y+h+SidebarBuffer)
	if lo >= hi {
		return plan
	}

	first := true
	add := func(start, end int) {
		if first {
			plan.Top = start
			first = false
		}
		plan.Bottom = end
	}

	pos, row := 0, 0
	for group, count := range counts {
		if count <= 0 {
			continue
		}
		if pos < hi && pos+SidebarHeaderHeight > lo {
			plan.Items = append(plan.Items, SidebarPlanItem{Group: group, Row: -1})
			add(pos, pos+SidebarHeaderHeight)
		}
		pos += SidebarHeaderHeight

		from := max(0, lo-pos) / SidebarRowHeight
		to := min(count, (max(0, hi-pos)+SidebarRowHeight-1)/SidebarRowHeight)
		if from < to {
			for i := from; i < to; i++ {
				plan.Items = append(plan.Items, SidebarPlanItem{Group: group, Row: row + i})
			}
			add(pos+from*SidebarRowHeight, pos+to*SidebarRowHeight)
		}
		pos += count * SidebarRowHeight
		row += count
	}
	return plan
}

type SidebarList struct {
	Pinned         []Thread
	Items          []SidebarItem
	Top            int
	Bottom         int
	Total          int
	Search         string
	SearchRev      int
	RenderID       int64
	ActiveThreadID string
}

type SidebarItem struct {
	Group  int
	Header bool
	Thread *Thread
}

func BuildSidebarList(threads []Thread, search string, bounds []time.Time, y, h int) SidebarList {
	query := strings.ToLower(strings.TrimSpace(search))
	var pinned, unpinned []Thread
	counts := make([]int, len(bounds)+1)
	for _, thread := range threads {
		if query != "" && !strings.Contains(strings.ToLower(thread.Title), query) {
			continue
		}
		if thread.IsPinned {
			pinned = append(pinned, thread)
			continue
		}
		unpinned = append(unpinned, thread)
		counts[threadGroup(thread.CreatedAt, bounds)]++
	}

	plan := PlanSidebar(counts, y, h)
	list := SidebarList{
		Pinned:   pinned,
		Top:      plan.Top,
		Bottom:   plan.Bottom,
		Total:    plan.Total,
		RenderID: time.Now().UnixNano(),
	}
	for _, item := range plan.Items {
		listItem := SidebarItem{Group: item.Group, Header: item.Row < 0}
		if item.Row >= 0 && item.Row < len(unpinned) {
			listItem.Thread = &unpinned[item.Row]
		}
		list.Items = append(list.Items, listItem)
	}

	return list
}

func threadGroup(createdAt time.Time, bounds []time.Time) int {
	for i, since := range bounds {
		if !createdAt.Before(since) {
			return i
		}
	}
	return len(bounds)
}

func threadGroupLabel(group int) string {
	if group < 0 || group >= len(threadGroupLabels) {
		return ""
	}
	return threadGroupLabels[group]
}
