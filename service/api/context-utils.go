// context-utils.go centralizes JSON helpers for reading request payloads and
// sending structured responses and errors.
package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// readJSON decodes the request body into dst and rejects unknown fields.
// Accepts Content-Type: application/json or application/json; charset=utf-8.
func readJSON(r *http.Request, dst interface{}) error {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		return http.ErrNotSupported
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// writeJSON writes value v as JSON with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeErrorResponse writes a minimal {code, message} JSON error payload.
func (rt *_router) writeErrorResponse(w http.ResponseWriter, status int, msg string) {
	resp := map[string]interface{}{
		"code":    status,
		"message": msg,
	}
	if err := writeJSON(w, status, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to write error response")
	}
}

// sendError is a compatibility wrapper over writeErrorResponse.
// The variadic parameter keeps call sites flexible; it is intentionally unused.
func (rt *_router) sendError(w http.ResponseWriter, status int, msg string, _ ...interface{}) {
	rt.writeErrorResponse(w, status, msg)
}
