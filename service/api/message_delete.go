package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// deleteMessage handles DELETE /messages/:id
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	messageID := ps.ByName("id")
	if messageID == "" {
		http.Error(w, `{"code":400,"message":"Message ID is required"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.DeleteMessage(ctx.UserID, messageID); err != nil {
		rt.baseLogger.WithError(err).Error("failed to delete message")
		http.Error(w, `{"code":403,"message":"Unauthorized or message not found"}`, http.StatusForbidden)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Message deleted",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode delete message response")
	}
}
