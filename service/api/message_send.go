package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage handles POST /messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		ToUserID  string `json:"toUserId"`
		ToGroupID string `json:"toGroupId"`
		Content   string `json:"content"`
	}

	// 解析请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode send request")
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 检查消息内容
	if req.Content == "" {
		http.Error(w, `{"code": 400, "message": "Message content cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// 构建消息对象
	message := models.Message{
		SenderID:   userID,
		Content:    req.Content,
		ReceiverID: req.ToUserID,
		GroupID:    req.ToGroupID,
	}

	var err error
	if req.ToUserID != "" {
		err = rt.db.SendPrivateMessage(message)
	} else if req.ToGroupID != "" {
		err = rt.db.SendGroupMessage(message)
	} else {
		http.Error(w, `{"code": 400, "message": "Must specify either toUserId or toGroupId"}`, http.StatusBadRequest)
		return
	}

	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to send message")
		http.Error(w, `{"code": 500, "message": "Failed to send message"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "Message sent successfully",
	}
	writeJSON(w, http.StatusCreated, resp)
}
