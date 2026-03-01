package server

import "net/http"

func (s *Server) appRoutes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := s.renderer.Render(w, "index", nil)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleNewMessage(w http.ResponseWriter, r *http.Request) {

}
