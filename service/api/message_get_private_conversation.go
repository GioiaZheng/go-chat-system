package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getPrivateConversation handles GET /messages/private?target_id=...
// English notes:
// - All errors use rt.sendError (no http.Error).
// - Success responses use writeJSON with a stable envelope.
func (rt *_router) getPrivateConversation(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetID := r.URL.Query().Get("target_id")
	if targetID == "" {
		rt.sendError(w, http.StatusBadRequest, "target_id is required")
		return
	}

	messages, err := rt.db.GetPrivateConversation(ctx.UserID, targetID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get private conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversation")
		return
	}

	// Ensure empty slice instead of null
	if messages == nil {
		messages = make([]models.Message, 0)
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode private conversation response")
	}
}
