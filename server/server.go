package server

import (
	"log"
	"log/slog"
	"net/http"

	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/views"

	"github.com/asaskevich/EventBus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	bus      EventBus.Bus
	router   chi.Router
	renderer *views.Renderer
	db       *db.Pool
	port     string
}

func StartServer() {
	conn, err := db.Connect(env.Env.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	s := &Server{
		bus:      EventBus.New(),
		router:   chi.NewRouter(),
		renderer: views.NewRenderer(env.Env.AssetVersion),
		db:       conn,
		port:     env.Env.Port,
	}

	s.router.
		With(middleware.Maybe(
			middleware.NoCache,
			func(*http.Request) bool { return env.Env.Dev },
		)).
		Get("/public/*",
			http.StripPrefix("/public/", http.FileServer(http.Dir("public"))).ServeHTTP,
		)

	s.router.Get("/", s.handleIndex)
	s.router.Get("/threads/{id}", handlePathVal(s.handleThread, "id"))
	s.router.Get("/threads/{id}/events", handlePathVal(s.handleThreadEvents, "id"))
	s.router.Post("/threads/{id}/archive", handlePathVal(s.handleArchiveThread, "id"))
	s.router.Get("/threads/search", handleSignals(s.handleSearchThreads))
	s.router.Post("/chat/send", handleSignals(s.handleSendMessage))

	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Fatal(err)
	}
}
