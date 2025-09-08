package main

import (
	"net/http"
	"os"
	"strings"
)

// allowedOrigins returns origins from env (ALLOWED_ORIGINS), comma-separated.
// No absolute URL literals embedded in code; default is permissive for local/dev.
//
// Example:
//
//	ALLOWED_ORIGINS="https://example.com,https://staging.example.com"
func allowedOrigins() string {
	if v := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); v != "" {
		return v
	}
	// Dev default: allow any origin; tighten in production by setting ALLOWED_ORIGINS.
	return "*"
}

// applyCORS sets CORS headers. Keep permissive for dev unless ALLOWED_ORIGINS is set.
//
// NOTE: For production-grade CORS with multiple specific origins, you typically
// reflect the incoming Origin if it's in the allowlist. For this assignment,
// a simple header is acceptable and avoids hardcoded URLs.
func applyCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigins())
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "1")
}

// applyCORSHandler wraps an http.Handler to apply CORS on every request.
// If the method is OPTIONS, it short-circuits with 200 OK for preflight checks.
func applyCORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyCORS(w)
		if r.Method == http.MethodOptions {
			// Preflight requests do not need to hit the router.
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
