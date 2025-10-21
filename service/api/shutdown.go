//go:build dev

package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

// POST /shutdown 仅开发使用（非 OpenAPI 的正式接口）
func (rt *_router) shutdown(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	const status = http.StatusOK
	resp := map[string]interface{}{"code": status, "message": "Shutdown requested"}
	_ = writeJSON(w, status, resp)
}
