package afastchat

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"afastchat.com/env"
)

//go:embed public/*
var publicFS embed.FS

func GetPublicHandler(env env.Env) http.Handler {
	embeddedPublic, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("sub public fs: %v", err)
	}
	publicAssets := http.FileServerFS(embeddedPublic)
	if env.Dev {
		publicAssets = http.FileServer(http.Dir("public"))
	}

	return http.StripPrefix("/public/", publicAssets)
}
