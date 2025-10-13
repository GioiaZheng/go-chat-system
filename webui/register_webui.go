// Package webui provides the optional Web UI wrapper for the API handler.
// It is referenced by cmd/webapi/register-web-ui.go when built with the "webui" build tag.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// Wrap mounts the embedded SPA under "/" and falls back to index.html for non-file paths.
// It returns a handler that first serves the SPA, then falls back to the given api handler.
func Wrap(api http.Handler) (http.Handler, error) {
	// Sub-FS rooted at "dist"
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	spa := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request looks like a static asset (has an extension), try to serve it from /dist.
		if hasExt(r.URL.Path) {
			spa.ServeHTTP(w, r)
			return
		}

		// For SPA routes (no extension), try index.html
		if r.Method == http.MethodGet {
			index, err := sub.Open("index.html")
			if err == nil {
				// Serve index.html content
				stat, _ := index.Stat()
				http.ServeContent(w, r, "index.html", stat.ModTime(), index)
				_ = index.Close()
				return
			}
		}

		// Otherwise fallback to API
		api.ServeHTTP(w, r)
	}), nil
}

func hasExt(p string) bool {
	ext := path.Ext(p)
	return ext != "" && strings.ContainsAny(ext, ".")
}
