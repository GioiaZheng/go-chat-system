package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupDetail handles GET /groups/:id
// English notes:
// - No http.Error; use rt.sendError for failures and writeJSON for success.
// - Log internal errors via ctx.Logger.
func (rt *_router) getGroupDetail(
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

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		// Not found is not necessarily an internal error; respond 404 without leaking details
		rt.sendError(w, http.StatusNotFound, "Group not found")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group retrieved",
		"data": map[string]interface{}{
			"group": group,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode get group response")
	}
}
