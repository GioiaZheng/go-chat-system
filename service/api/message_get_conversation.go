package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessages handles GET /messages?chat_type=private|group&target_id=...
// English notes:
// - Validate query params; return consistent envelopes.
// - Use rt.sendError + logger for errors.
func (rt *_router) getMessages(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chatType := r.URL.Query().Get("chat_type")
	targetID := r.URL.Query().Get("target_id")
	if chatType == "" || targetID == "" {
		rt.sendError(w, http.StatusBadRequest, "chat_type and target_id are required")
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
		rt.sendError(w, http.StatusBadRequest, "chat_type must be 'private' or 'group'")
		return
	}

	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversation")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode conversation response")
	}
}
