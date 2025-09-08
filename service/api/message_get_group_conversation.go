package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupConversation handles GET /messages/group?target_id=...
// English notes:
// - Validate target_id; return consistent envelopes.
// - Use rt.sendError + logger for errors.
func (rt *_router) getGroupConversation(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groupID := r.URL.Query().Get("target_id")
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "target_id is required")
		return
	}

	messages, err := rt.db.GetGroupConversation(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get group conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversation")
		return
	}
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
		ctx.Logger.WithError(err).Error("failed to encode group conversation response")
	}
}
