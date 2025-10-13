//go:build webui
// +build webui

package webui

import (
	"embed"
	"net/http"
)

//go:embed dist/*
var distFS embed.FS

// Register wraps the API handler with a file server for the SPA assets.
func Register(api http.Handler) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(distFS)))
	// 如果想把 /api/* 代理到原 API，可以把下一行打开
	// mux.Handle("/api/", http.StripPrefix("/api", api))
	return mux, nil
}
