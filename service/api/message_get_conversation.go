package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessages handles GET /messages?chat_type=private|group&target_id=...
// It dispatches to private or group conversation fetchers and guarantees [] instead of null.
func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	chatType := r.URL.Query().Get("chat_type")
	targetID := r.URL.Query().Get("target_id")
	if chatType == "" || targetID == "" {
		http.Error(w, `{"code":400,"message":"chat_type and target_id are required"}`, http.StatusBadRequest)
		return
	}

	var (
		messages interface{}
		err      error
	)

	switch chatType {
	case "private":
		var list []models.Message
		list, err = rt.db.GetPrivateConversation(ctx.UserID, targetID)
		if err == nil && list == nil {
			list = make([]models.Message, 0)
		}
		messages = list
	case "group":
		var list []models.Message
		list, err = rt.db.GetGroupConversation(targetID)
		if err == nil && list == nil {
			list = make([]models.Message, 0)
		}
		messages = list
	default:
		http.Error(w, `{"code":400,"message":"chat_type must be 'private' or 'group'"}`, http.StatusBadRequest)
		return
	}

	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get conversation")
		http.Error(w, `{"code":500,"message":"Failed to fetch conversation"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode conversation response")
	}
}
