package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// uncommentMessage handles DELETE /messages/:id/comment
// English notes:
// - All errors via rt.sendError; success via writeJSON.
// - Logs internal DB failures via ctx.Logger.
func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	messageID := ps.ByName("id")
	if messageID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing message ID")
		return
	}

	if err := rt.db.UncommentMessage(messageID); err != nil {
		ctx.Logger.WithError(err).Error("failed to uncomment message")
		rt.sendError(w, http.StatusInternalServerError, "Failed to uncomment message")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Comment removed successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode uncomment response")
	}
}
