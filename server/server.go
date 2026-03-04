package server

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"datastar-chat/db"
)

type Server struct {
	mux      *http.ServeMux
	renderer *TemplateRenderer
	db       *db.Queries
	streams  *StreamRegistry
}

func New(assetVersion string, dev bool) *Server {
	queries, err := db.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	s := &Server{
		mux:      http.NewServeMux(),
		renderer: newRenderer(assetVersion, dev),
		db:       queries,
		streams:  newStreamRegistry(),
	}

	s.mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	s.appRoutes()

	return s
}

func (s *Server) Start() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		log.Fatal(err)
	}
}
