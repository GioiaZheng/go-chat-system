package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type ForwardRequest struct {
	// English notes:
	// Keep current JSON fields toUserId/toGroupId for client compatibility.
	// If later your OpenAPI changes, adapt tags accordingly (or accept both).
	ToUserID  string `json:"toUserId,omitempty"`
	ToGroupID string `json:"toGroupId,omitempty"`
}

// forwardMessage handles POST /messages/:id/forward
// English notes:
// - Validate exactly one target: either user or group.
// - Use rt.sendError on failure; writeJSON on success.
func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing message id")
		return
	}
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ForwardRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if (req.ToUserID == "" && req.ToGroupID == "") || (req.ToUserID != "" && req.ToGroupID != "") {
		rt.sendError(w, http.StatusBadRequest, "must provide either toUserId or toGroupId")
		return
	}

	if err := rt.db.ForwardMessage(ctx.UserID, msgID, req.ToUserID, req.ToGroupID); err != nil {
		ctx.Logger.WithError(err).Error("failed to forward message")
		rt.sendError(w, http.StatusInternalServerError, "failed to forward message")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message forwarded",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode forward response")
	}
}
