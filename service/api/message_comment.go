package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type CommentRequest struct {
	Comment string `json:"comment"`
}

// commentMessage handles POST /messages/:id/comment
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	if messageID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing message id")
		return
	}
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CommentRequest
	if err := readJSON(r, &req); err != nil || req.Comment == "" {
		rt.sendError(w, http.StatusBadRequest, "comment is required")
		return
	}

	if err := rt.db.CommentMessage(messageID, ctx.UserID, req.Comment); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to comment message")
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Comment added",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode comment response")
	}
}
