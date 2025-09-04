package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type AddMemberRequest struct {
	UserID string `json:"userId"`
}

// addToGroup handles POST /groups/:id/add
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddMemberRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		rt.sendError(w, http.StatusBadRequest, "userId is required")
		return
	}

	// Use bulk API with a single member for simplicity
	if err := rt.db.AddGroupMembers(groupID, []string{req.UserID}); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to add member to group")
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Member added to group",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode add-to-group response")
	}
}
