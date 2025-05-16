package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness is an HTTP handler that checks the API server status.
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":200,"message":"Service is alive"}`))
}
