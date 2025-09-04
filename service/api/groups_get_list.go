package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupsList handles GET /groups
// OpenAPI: response uses top-level "items" for the collection.
func (rt *_router) getGroupsList(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	groups, err := rt.db.GetGroupsList(userID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to get groups list"}`, http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = make([]models.Group, 0)
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Groups list retrieved",
		"items":   groups,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode groups list response")
	}
}
