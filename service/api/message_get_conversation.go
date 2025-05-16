package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getConversation handles GET /messages
// 只从查询参数(Query Parameters)获取 chat_type 和 target_id
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 从URL Query取参数
	query := r.URL.Query()
	chatType := query.Get("chat_type") // private 或 group
	targetID := query.Get("target_id") // 目标用户ID 或 群组ID

	if chatType == "" || targetID == "" {
		http.Error(w, `{"code": 400, "message": "chat_type and target_id are required"}`, http.StatusBadRequest)
		return
	}

	var messages []models.Message
	var err error

	if chatType == "private" {
		// 私聊查询
		messages, err = rt.db.GetPrivateConversation(userID, targetID)
	} else if chatType == "group" {
		// 群聊查询
		messages, err = rt.db.GetGroupConversation(targetID)
	} else {
		http.Error(w, `{"code": 400, "message": "Invalid chat_type: must be 'private' or 'group'"}`, http.StatusBadRequest)
		return
	}

	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get conversation")
		http.Error(w, `{"code": 500, "message": "Failed to get conversation"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}
