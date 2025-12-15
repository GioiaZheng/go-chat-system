//go:build webui

// register-web-ui.go mounts the compiled frontend assets when the binary is
// built with the `webui` tag, while forwarding API paths to the existing API
// handler.
package main

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

// registerWebUI mounts the embedded frontend assets at the root path when the
// webui build tag is enabled, while delegating API routes back to the provided
// handler. This keeps routing concerns localized to a single place.
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
