package server

import (
	"log"
	"log/slog"
	"net/http"

	"datastar-chat/db"
)

type Server struct {
	mux      *http.ServeMux
	renderer *TemplateRenderer
	db       *db.Queries
	streams  *StreamRegistry
	port     string
}

func New(databaseURL, port, assetVersion string, dev bool) *Server {
	queries, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	s := &Server{
		mux:      http.NewServeMux(),
		renderer: newRenderer(assetVersion, dev),
		db:       queries,
		streams:  newStreamRegistry(),
		port:     port,
	}

	s.mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	s.appRoutes()

	return s
}

func (s *Server) Start() {
	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		log.Fatal(err)
	}
}
