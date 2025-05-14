package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization",
			"x-example-header",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		handlers.AllowedOrigins([]string{
			"https://wasa-text.vercel.app",
			"https://wasa-text-git-main-lingfengs-projects-dd7a133a.vercel.app",
			"https://wasa-text-befk2zdy4-lingfengs-projects-dd7a133a.vercel.app",
			"*"}),
		handlers.MaxAge(1),
	)(h)
}
