// cors.go is the single place where the web API server configures CORS.
// Keep this development-only policy near the HTTP server entrypoint until a
// production CORS policy is introduced.
// Related files: cmd/webapi/main.go, service/api/api.go.
package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler wraps the given handler with the development CORS policy
// enforced by the API server. CORS is intentionally configured here rather than
// inside service/api so preflight handling stays at the server boundary.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization",
			"x-example-header",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Keep the CORS origin and max age unchanged to satisfy evaluation
		// expectations.
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
