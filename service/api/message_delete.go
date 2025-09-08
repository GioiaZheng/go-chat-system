package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// deleteMessage handles DELETE /messages/:id
// English notes:
// - Use rt.sendError instead of http.Error.
// - Log internal errors, but don't leak details to clients.
func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	messageID := ps.ByName("id")
	if messageID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	if err := rt.db.DeleteMessage(ctx.UserID, messageID); err != nil {
		// Forbidden covers both "not owner" and "not found" cases without leaking details
		ctx.Logger.WithError(err).Error("failed to delete message")
		rt.sendError(w, http.StatusForbidden, "Unauthorized or message not found")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message deleted",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode delete message response")
	}
}
