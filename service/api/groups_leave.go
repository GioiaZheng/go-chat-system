package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// leaveGroup handles DELETE /groups/:id/members
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("id")
	userID := GetUserIDFromContext(r.Context())

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := rt.db.LeaveGroup(groupID, userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to leave group")
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":    200,
		"message": "Left group successfully",
	})
}
