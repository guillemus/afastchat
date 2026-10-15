package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"regexp"

	"afastchat.com/views"
	"afastchat.com/views/ui"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

const maxSidebarSearchRunes = 200

var sidebarTabPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

type sidebarView struct {
	Tab       string  `json:"tab"`
	Y         float64 `json:"y"`
	H         float64 `json:"h"`
	Search    string  `json:"search"`
	SearchRev int     `json:"searchRev"`
}

func readSidebarView(w http.ResponseWriter, r *http.Request) (sidebarView, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var sig struct {
		Sidebar sidebarView `json:"sidebar"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		return sidebarView{}, err
	}
	if !sidebarTabPattern.MatchString(sig.Sidebar.Tab) {
		return sidebarView{}, errors.New("invalid sidebar tab")
	}
	return sig.Sidebar, nil
}

func sidebarPixels(v float64) int {
	if math.IsNaN(v) {
		return 0
	}
	return int(math.Round(min(max(v, -1e9), 1e9)))
}

func sidebarSearchQuery(search string) string {
	runes := []rune(search)
	if len(runes) > maxSidebarSearchRunes {
		return string(runes[:maxSidebarSearchRunes])
	}
	return search
}

func (x *Server) sidebarList(
	ctx context.Context, userID string, view sidebarView, activeThreadID string,
) (views.SidebarList, error) {
	threads, err := x.db.ListThreads(ctx, userID)
	if err != nil {
		return views.SidebarList{}, fmt.Errorf("list threads: %w", err)
	}

	list := views.BuildSidebarList(
		mapViewThreads(threads),
		sidebarSearchQuery(view.Search),
		views.ThreadGroupBounds(ctx),
		sidebarPixels(view.Y),
		sidebarPixels(view.H),
	)
	list.Search = view.Search
	list.SearchRev = view.SearchRev
	list.ActiveThreadID = activeThreadID

	return list, nil
}

func (x *Server) initialSidebarList(ctx context.Context, userID, activeThreadID string) (views.SidebarList, error) {
	return x.sidebarList(ctx, userID, sidebarView{H: views.SidebarDefaultHeight}, activeThreadID)
}

func (x *Server) handleSidebarEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	view, err := readSidebarView(w, r)
	if err != nil {
		handleBadReq(w, "read sidebar signals", err)
		return
	}

	viewSubject := SidebarSubject(user.ID, view.Tab)
	messages := make(chan *nats.Msg, 64)
	unsubscribe, err := x.broker.SubscribeSubjects(messages, UserThreadsSubject(user.ID), viewSubject)
	if err != nil {
		slog.Error("subscribe to sidebar events", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer unsubscribe()

	apply := func(msg *nats.Msg) {
		if msg.Subject != viewSubject {
			return
		}
		var next sidebarView
		if err := json.Unmarshal(msg.Data, &next); err != nil {
			slog.Error("decode sidebar view", "err", err)
			return
		}
		next.Tab = view.Tab
		view = next
	}

	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli(
		datastar.WithBrotliLGWin(20),
		datastar.WithBrotliLevel(3),
	)))
	failed := false
	for {
		list, err := x.sidebarList(ctx, user.ID, view, "")
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			slog.Error("load sidebar", "err", err)
			if !failed {
				patchToastOnSSE(sse, ui.Toast{Kind: ui.ToastError, Message: "We couldn't refresh your chats. Refresh the page."})
			}
			failed = true
		default:
			failed = false
			views.PatchSidebar(sse, list)
		}

		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			apply(msg)
		}
	drain:
		for {
			select {
			case msg := <-messages:
				apply(msg)
			default:
				break drain
			}
		}
	}
}

func (x *Server) handleSidebarView(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	view, err := readSidebarView(w, r)
	if err != nil {
		handleBadReq(w, "read sidebar view signals", err)
		return
	}

	x.broker.PublishSidebarView(user.ID, view)
	w.WriteHeader(http.StatusNoContent)
}
