package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness is an HTTP handler that checks the API server status.
// If the server cannot serve requests (e.g., database not ready), it replies with HTTP Status 500.
// Otherwise, it replies with HTTP Status 200 and a JSON body {"status":"ok"}
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := rt.db.Ping(); err != nil {
		rt.baseLogger.WithError(err).Error("database ping failed")
		http.Error(w, "database not ready", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
