package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// ForwardMessageRequest 转发请求体
type ForwardMessageRequest struct {
	ToUserID  string `json:"toUserId,omitempty"`
	ToGroupID string `json:"toGroupId,omitempty"`
}

// forwardMessage handles POST /messages/:id/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("id")
	if messageID == "" {
		http.Error(w, `{"code": 400, "message": "Message ID is required"}`, http.StatusBadRequest)
		return
	}

	var req ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode forward request")
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 确保目标用户或群组存在
	if req.ToUserID == "" && req.ToGroupID == "" {
		http.Error(w, `{"code": 400, "message": "Must specify either toUserId or toGroupId"}`, http.StatusBadRequest)
		return
	}

	// 调用数据库方法转发消息
	err := rt.db.ForwardMessage(userID, messageID, req.ToUserID, req.ToGroupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to forward message")
		http.Error(w, `{"code": 500, "message": "Failed to forward message"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response := map[string]interface{}{
		"code":    200,
		"message": "Message forwarded successfully",
	}
	writeJSON(w, http.StatusOK, response)
}
