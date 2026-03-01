package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Server struct {
	mux      *http.ServeMux
	renderer *TemplateRenderer
}

func New(assetVersion string, dev bool) *Server {
	s := &Server{
		mux:      http.NewServeMux(),
		renderer: newRenderer(assetVersion, dev),
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
	fmt.Printf("listening on %s\n", addr)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		log.Fatal(err)
	}
}
