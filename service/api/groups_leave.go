package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// leaveGroup handles POST /groups/:id/leave
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		http.Error(w, `{"code":400,"message":"Group ID is required"}`, http.StatusBadRequest)
		return
	}
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := rt.db.LeaveGroup(groupID, ctx.UserID); err != nil {
		rt.baseLogger.WithError(err).Error("failed to leave group")
		http.Error(w, `{"code":500,"message":"Failed to leave group"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Left group successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode leave group response")
	}
}
