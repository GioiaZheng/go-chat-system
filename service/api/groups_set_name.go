package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type UpdateGroupNameRequest struct {
	Name string `json:"name"`
}

// setGroupName handles PUT /groups/:id/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateGroupNameRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		rt.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to update group name")
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Group name updated",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode set group name response")
	}
}
