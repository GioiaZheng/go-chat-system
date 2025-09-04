package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// shutdown handles POST /shutdown (admin/dev only)
// NOTE: For assignments, we keep it simple: return 200 OK.
func (rt *_router) shutdown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	resp := map[string]interface{}{
		"code":    200,
		"message": "Shutdown requested",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode shutdown response")
	}
}
