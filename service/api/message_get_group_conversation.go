package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupConversation handles GET /messages/group?target_id=...
func (rt *_router) getGroupConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("target_id")
	if groupID == "" {
		http.Error(w, `{"code":400,"message":"target_id is required"}`, http.StatusBadRequest)
		return
	}

	messages, err := rt.db.GetGroupConversation(groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get group conversation")
		http.Error(w, `{"code":500,"message":"Failed to fetch conversation"}`, http.StatusInternalServerError)
		return
	}
	// Ensure empty slice instead of null
	if messages == nil {
		messages = make([]models.Message, 0)
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode group conversation response")
	}

}
