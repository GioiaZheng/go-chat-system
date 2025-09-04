package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness reports that the API is up.
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	resp := map[string]interface{}{
		"code":    200,
		"message": "Service is alive",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		// Best-effort fallback; at least log the failure
		rt.baseLogger.WithError(err).Error("failed to write liveness response")
	}
}
