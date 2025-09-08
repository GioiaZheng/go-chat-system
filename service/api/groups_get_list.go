package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupsList handles GET /groups
// English notes:
// - On error, use rt.sendError.
// - On success, respond with an "items" array (empty array if no groups).
func (rt *_router) getGroupsList(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := ctx.UserID
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groups, err := rt.db.GetGroupsList(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get groups list")
		rt.sendError(w, http.StatusInternalServerError, "Failed to get groups list")
		return
	}
	if groups == nil {
		groups = make([]models.Group, 0)
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Groups list retrieved",
		"items":   groups,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode groups list response")
	}
}
