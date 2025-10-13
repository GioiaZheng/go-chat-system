package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// shutdown handles POST /shutdown (development / admin only).
//
// This endpoint is NOT part of the OpenAPI specification and should not
// be exposed in production builds. It simply returns 200 OK so that
// automated scripts or local testing can confirm graceful termination.
func (rt *_router) shutdown(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	const status = http.StatusOK

	resp := map[string]interface{}{
		"code":    status,
		"message": "Shutdown requested",
	}

	if err := writeJSON(w, status, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode shutdown response")
	}
}
