package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupDetail handles GET /groups/:id
func (rt *_router) getGroupDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		http.Error(w, `{"code":400,"message":"Group ID is required"}`, http.StatusBadRequest)
		return
	}
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		http.Error(w, `{"code":404,"message":"Group not found"}`, http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Group retrieved",
		"data": map[string]interface{}{
			"group": group,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode get group response")
	}
}
