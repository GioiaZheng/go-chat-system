package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Handler returns an http.Handler that serves embedded frontend assets.
// It also falls back to index.html for unknown routes (typical SPA behavior).
func Handler() http.Handler {
	// Serve static files from /dist
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// If embedding is broken, return 404 handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	}
	fileSrv := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Normalize request path
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		// Try to serve a static asset under /dist first
		if p == "" || p == "." {
			p = "index.html"
		}
		if f, err := sub.Open(p); err == nil {
			_ = f.Close()
			fileSrv.ServeHTTP(w, r)
			return
		}

		// Fallback: serve index.html for SPA routes
		data, err := fs.ReadFile(distFS, "dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}
