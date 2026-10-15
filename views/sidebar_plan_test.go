package views

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

type naiveItem struct {
	item       SidebarPlanItem
	start, end int
}

func naiveLayout(counts []int) []naiveItem {
	var items []naiveItem
	pos, row := 0, 0
	for group, count := range counts {
		if count <= 0 {
			continue
		}
		items = append(items, naiveItem{SidebarPlanItem{Group: group, Row: -1}, pos, pos + SidebarHeaderHeight})
		pos += SidebarHeaderHeight
		for i := range count {
			items = append(items, naiveItem{SidebarPlanItem{Group: group, Row: row + i}, pos, pos + SidebarRowHeight})
			pos += SidebarRowHeight
		}
		row += count
	}
	return items
}

func naivePlan(counts []int, y, h int) SidebarPlan {
	y, h = clampSidebarView(y, h)
	layout := naiveLayout(counts)
	var plan SidebarPlan
	if len(layout) > 0 {
		plan.Total = layout[len(layout)-1].end
	}
	lo, hi := max(0, y-SidebarBuffer), min(plan.Total, y+h+SidebarBuffer)
	if lo >= hi {
		return plan
	}
	for _, it := range layout {
		if it.start >= hi || it.end <= lo {
			continue
		}
		if len(plan.Items) == 0 {
			plan.Top = it.start
		}
		plan.Bottom = it.end
		plan.Items = append(plan.Items, it.item)
	}
	return plan
}

func needsFetch(plan SidebarPlan, y, h int) bool {
	h = min(max(h, 0), SidebarMaxHeight)
	lo, hi := max(0, y-SidebarBuffer/2), min(plan.Total, y+h+SidebarBuffer/2)
	return lo < hi && (lo < plan.Top || hi > plan.Bottom)
}

func TestPlanSidebarMatchesNaive(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		counts := make([]int, 5)
		for i := range counts {
			switch rng.IntN(4) {
			case 0:
			case 1:
				counts[i] = rng.IntN(3)
			default:
				counts[i] = rng.IntN(400)
			}
		}
		total := naivePlan(counts, 0, 0).Total
		y := rng.IntN(total+4000) - 2000
		h := rng.IntN(3000)
		if rng.IntN(20) == 0 {
			h = rng.IntN(100000)
		}

		got := PlanSidebar(counts, y, h)
		want := naivePlan(counts, y, h)
		if got.Top != want.Top ||
			got.Bottom != want.Bottom || got.Total != want.Total || !slices.Equal(got.Items, want.Items) {
			t.Fatalf("counts=%v y=%d h=%d\ngot  %+v\nwant %+v", counts, y, h, got, want)
		}
		if needsFetch(got, y, h) {
			t.Fatalf("plan does not cover its own view: counts=%v y=%d h=%d plan=%+v", counts, y, h, got)
		}
	}
}

func TestPlanSidebarEdges(t *testing.T) {
	tests := []struct {
		name   string
		counts []int
		y, h   int
		want   SidebarPlan
	}{
		{name: "no threads", counts: []int{0, 0, 0, 0, 0}, y: 0, h: 800, want: SidebarPlan{}},
		{name: "nil counts", y: 0, h: 800, want: SidebarPlan{}},
		{
			name:   "single thread",
			counts: []int{0, 1, 0, 0, 0},
			y:      0, h: 800,
			want: SidebarPlan{
				Top: 0, Bottom: 71, Total: 71,
				Items: []SidebarPlanItem{{Group: 1, Row: -1}, {Group: 1, Row: 0}},
			},
		},
		{name: "viewport above list", counts: []int{10}, y: -5000, h: 800, want: SidebarPlan{Total: 350}},
		{name: "viewport below list", counts: []int{10}, y: 5000, h: 800, want: SidebarPlan{Total: 350}},
		{name: "zero height still buffers", counts: []int{1000}, y: 0, h: 0, want: naivePlan([]int{1000}, 0, 0)},
		{name: "negative height", counts: []int{1000}, y: 100, h: -50, want: naivePlan([]int{1000}, 100, 0)},
		{name: "huge y", counts: []int{1000}, y: 1 << 60, h: 800, want: SidebarPlan{Total: 31040}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanSidebar(tt.counts, tt.y, tt.h)
			if got.Top != tt.want.Top ||
				got.Bottom != tt.want.Bottom || got.Total != tt.want.Total || !slices.Equal(got.Items, tt.want.Items) {
				t.Fatalf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestPlanSidebarWindowSize(t *testing.T) {
	plan := PlanSidebar([]int{100000}, 1_000_000, 900)
	if len(plan.Items) > (900+2*SidebarBuffer)/SidebarRowHeight+2 {
		t.Fatalf("window too large: %d rows", len(plan.Items))
	}
	plan = PlanSidebar([]int{100000}, 0, 1<<40)
	if len(plan.Items) > (SidebarMaxHeight+SidebarBuffer)/SidebarRowHeight+2 {
		t.Fatalf("height not clamped: %d rows", len(plan.Items))
	}
}

func TestBuildSidebarList(t *testing.T) {
	loc := time.FixedZone("viewer", 2*60*60)
	today := time.Date(2026, 10, 15, 0, 0, 0, 0, loc)
	bounds := []time.Time{today, today.AddDate(0, 0, -1), today.AddDate(0, 0, -7), today.AddDate(0, 0, -30)}
	at := func(days, hours int) time.Time {
		return today.AddDate(0, 0, -days).Add(time.Duration(hours) * time.Hour).UTC()
	}
	threads := []Thread{
		{ID: "pinned", Title: "Pinned Über", IsPinned: true, CreatedAt: at(40, 0)},
		{ID: "a", Title: "ÜBER today", CreatedAt: at(0, 10)},
		{ID: "b", Title: "midnight", CreatedAt: at(0, 0)},
		{ID: "c", Title: "just before midnight über", CreatedAt: at(0, 0).Add(-time.Second)},
		{ID: "d", Title: "last week", CreatedAt: at(3, 0)},
		{ID: "e", Title: "ancient über", CreatedAt: at(400, 0)},
	}
	ids := func(list SidebarList) []string {
		var out []string
		for _, thread := range list.Pinned {
			out = append(out, "*"+thread.ID)
		}
		for _, item := range list.Items {
			switch {
			case item.Header:
				out = append(out, threadGroupLabel(item.Group))
			case item.Thread == nil:
				out = append(out, "missing")
			default:
				out = append(out, item.Thread.ID)
			}
		}
		return out
	}

	got := ids(BuildSidebarList(threads, "", bounds, 0, 800))
	want := []string{"*pinned", "Today", "a", "b", "Yesterday", "c", "Previous 7 days", "d", "Older", "e"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	got = ids(BuildSidebarList(threads, "  über ", bounds, 0, 800))
	want = []string{"*pinned", "Today", "a", "Yesterday", "c", "Older", "e"}
	if !slices.Equal(got, want) {
		t.Fatalf("search: got %v want %v", got, want)
	}

	list := BuildSidebarList(threads, "nothing matches", bounds, 0, 800)
	if len(list.Pinned) != 0 || len(list.Items) != 0 || list.Total != 0 {
		t.Fatalf("expected empty list, got %+v", list)
	}
}
