package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness responds with a minimal JSON envelope indicating the service is up.
// OpenAPI: GET /liveness -> { code: 200, message: "Service is alive" }.
func (rt *_router) liveness(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	const status = http.StatusOK

	resp := map[string]interface{}{
		"code":    status,
		"message": "Service is alive",
	}

	if err := writeJSON(w, status, resp); err != nil {
		// Best-effort: log the failure so it shows up in server logs.
		rt.baseLogger.WithError(err).Error("failed to write liveness response")
	}
}
