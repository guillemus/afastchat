package afastchat

import (
	"embed"
	"io/fs"
	"net/http"

	"afastchat.com/config"
)

//go:embed public/*
var publicFS embed.FS

func PublicHandler() http.Handler {
	var files http.Handler
	if config.Dev {
		files = http.StripPrefix("/public/", http.FileServer(http.Dir("public")))
	} else {
		public, err := fs.Sub(publicFS, "public")
		if err != nil {
			panic("create public filesystem: " + err.Error())
		}
		files = http.StripPrefix("/public/", http.FileServerFS(public))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Dev {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}
