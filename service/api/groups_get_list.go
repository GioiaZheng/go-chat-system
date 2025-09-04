package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupsList handles GET /groups
func (rt *_router) getGroupsList(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Fetch groups the user belongs to
	groups, err := rt.db.GetGroupsList(userID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to get groups list"}`, http.StatusInternalServerError)
		return
	}
	// Ensure empty slice instead of null
	if groups == nil {
		groups = make([]models.Group, 0)
	}

	// Build response payload
	resp := map[string]interface{}{
		"code":    200,
		"message": "Groups list retrieved successfully",
		"data": map[string]interface{}{
			"groups": groups,
		},
	}

	// Use centralized writeJSON and handle error
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode groups list response")
	}
}
