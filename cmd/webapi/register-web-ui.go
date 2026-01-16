//go:build webui

// register-web-ui.go mounts the compiled frontend assets when the binary is
// built with the `webui` tag, while forwarding API paths to the existing API
// handler.
// Related files: cmd/webapi/register-web-ui-stub.go, webui/register_webui.go.
package main

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

// registerWebUI mounts the embedded frontend assets at the root path when the
// webui build tag is enabled and delegates API routes to the provided handler.
// This keeps routing concerns localized to a single place.
func registerWebUI(api http.Handler) (http.Handler, error) {
	mux := http.NewServeMux()

	// Serve the SPA entrypoint.
	mux.Handle("/", webui.Handler())

	// Forward API routes to the backend handler. If the API lives at the root,
	// swap the order: register prefixes on api first, then hand "/" to the
	// frontend.
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
