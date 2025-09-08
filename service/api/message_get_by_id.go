package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessageById handles GET /messages/:id
// English notes:
// - Use rt.sendError for all failures; log internal errors.
// - Return consistent envelope on success.
func (rt *_router) getMessageById(
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

	message, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get message by ID")
		rt.sendError(w, http.StatusNotFound, "Message not found")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message fetched successfully",
		"data": map[string]interface{}{
			"message": message,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode message response")
	}
}
