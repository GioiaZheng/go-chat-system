package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type ForwardRequest struct {
	ToUserID  string `json:"toUserId,omitempty"`
	ToGroupID string `json:"toGroupId,omitempty"`
}

// forwardMessage handles POST /messages/:id/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	msgID := ps.ByName("id")
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing message id")
		return
	}
	if ctx.UserID == "" {
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
		rt.sendError(w, http.StatusInternalServerError, "failed to forward message")
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Message forwarded",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode forward response")
	}
}
