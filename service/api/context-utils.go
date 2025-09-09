package api

import (
	"encoding/json"
	"net/http"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
)

// readJSON decodes request body into dst, rejecting unknown fields.
func readJSON(r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" && r.Header.Get("Content-Type") != "application/json" {
		return http.ErrNotSupported
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeErrorResponse writes {code,message} as JSON.
func (rt *_router) writeErrorResponse(w http.ResponseWriter, status int, msg string) {
	resp := map[string]interface{}{
		"code":    status,
		"message": msg,
	}
	if err := writeJSON(w, status, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to write error response")
	}
}

// sendError is a compatibility wrapper.
// It accepts an optional RequestContext (ignored if not provided).
func (rt *_router) sendError(w http.ResponseWriter, status int, msg string, _ ...reqcontext.RequestContext) {
	rt.writeErrorResponse(w, status, msg)
}
