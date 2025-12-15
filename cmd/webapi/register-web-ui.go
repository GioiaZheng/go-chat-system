//go:build webui

package main

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

// When built with -tags webui, mount the embedded frontend assets at "/"
// and route all remaining paths to the API handler.
func registerWebUI(api http.Handler) (http.Handler, error) {
	mux := http.NewServeMux()

        // Frontend entrypoint
        mux.Handle("/", webui.Handler())

        // API routes (delegate common prefixes to the existing API handler)
        // If your API lives at the root, swap the order: register prefixes
        // on api first, then hand "/" to the frontend.
	mux.Handle("/session", api)
	mux.Handle("/liveness", api)
	mux.Handle("/users/", api)
	mux.Handle("/groups/", api)
	mux.Handle("/messages/", api)
	mux.Handle("/conversations", api)
	mux.Handle("/conversations/", api)
	mux.Handle("/uploads/", api)

	return mux, nil
}
