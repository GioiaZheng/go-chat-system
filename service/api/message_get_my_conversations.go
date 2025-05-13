package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations handles GET /messages/conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 获取用户的会话列表
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get my conversations")
		http.Error(w, `{"code": 500, "message": "Failed to fetch conversations"}`, http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"code":    200,
		"message": "Conversations fetched successfully",
		"data": map[string]interface{}{
			"conversations": conversations,
		},
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode conversations response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
