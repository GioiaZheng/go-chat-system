package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// leaveGroup handles POST /groups/:id/leave
// English notes:
// - Use rt.sendError for all failures, log internal errors.
// - Keep response shape consistent project-wide.
func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := ps.ByName("id")
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Group ID is required")
		return
	}
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := rt.db.LeaveGroup(groupID, ctx.UserID); err != nil {
		ctx.Logger.WithError(err).Error("failed to leave group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to leave group")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Left group successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode leave group response")
	}
}
