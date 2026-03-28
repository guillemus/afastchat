package server

import (
	"log"
	"log/slog"
	"net/http"

	"datastar-chat/db"
	"datastar-chat/views"
)

type Server struct {
	mux      *http.ServeMux
	renderer *views.Renderer
	db       *db.Queries
	streams  *StreamRegistry
	port     string
}

func New(databaseURL, port, assetVersion string, _ bool) *Server {
	queries, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	s := &Server{
		mux:      http.NewServeMux(),
		renderer: views.NewRenderer(assetVersion),
		db:       queries,
		streams:  newStreamRegistry(),
		port:     port,
	}

	s.mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /threads/{id}", s.handleThread)
	s.mux.HandleFunc("GET /threads/{id}/stream", s.handleResumeStream)
	s.mux.HandleFunc("GET /threads/search", s.handleSearchThreads)
	s.mux.HandleFunc("POST /chat/send", s.handleSendMessage)
	s.mux.HandleFunc("POST /threads/{id}/archive", s.handleArchiveThread)

	return s
}

func (s *Server) Start() {
	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		log.Fatal(err)
	}
}
