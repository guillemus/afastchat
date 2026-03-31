package server

import (
	"log"
	"log/slog"
	"net/http"

	"datastar-chat/db"
	"datastar-chat/env"
	"datastar-chat/views"
)

type Server struct {
	mux      *http.ServeMux
	renderer *views.Renderer
	db       *db.Conn
	port     string
}

func StartServer() {
	conn, err := db.Connect(env.Env.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	s := &Server{
		mux:      http.NewServeMux(),
		renderer: views.NewRenderer(env.Env.AssetVersion),
		db:       conn,
		port:     env.Env.Port,
	}

	s.mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /threads/{id}", handleUUIDPathVal(s.handleThread, "id"))
	s.mux.HandleFunc("GET /threads/search", handleSignals(s.handleSearchThreads))
	s.mux.HandleFunc("POST /chat/send", handleSignals(s.handleSendMessage))
	s.mux.HandleFunc("POST /threads/{id}/archive", handleUUIDPathVal(s.handleArchiveThread, "id"))

	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		log.Fatal(err)
	}
}
