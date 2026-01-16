// shutdown.go provides a dev-only POST /shutdown hook so integration tests can
// stop the server gracefully without resorting to process signals.
// Related files: service/api/api-handler.go, cmd/webapi/main.go.
//go:build dev

package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// POST /shutdown is for development only (not part of the formal OpenAPI).
func (rt *_router) shutdown(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	const status = http.StatusOK
	resp := map[string]interface{}{"code": status, "message": "Shutdown requested"}
	_ = writeJSON(w, status, resp)
}
