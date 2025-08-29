package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// CommentMessageRequest 评论消息请求体
type CommentMessageRequest struct {
	Comment string `json:"comment"`
}

// commentMessage 处理 POST /messages/:messageId/comment
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	userID := ctx.UserID

	// 校验 messageId
	idPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if messageID == "" || !idPattern.MatchString(messageID) || len(messageID) > 64 {
		http.Error(w, `{"code": 400, "message": "Invalid message ID format"}`, http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req CommentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验评论内容
	commentPattern := regexp.MustCompile(`^[a-zA-Z0-9\s.,!?]+$`)
	if req.Comment == "" || !commentPattern.MatchString(req.Comment) || len(req.Comment) > 1000 {
		http.Error(w, `{"code": 400, "message": "Invalid comment: must be 1-1000 characters long and match the pattern [a-zA-Z0-9\\s.,!?]+"}`, http.StatusBadRequest)
		return
	}

	// 添加评论
	if err := rt.db.CommentMessage(messageID, userID, req.Comment); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to add comment to message")
		http.Error(w, `{"code": 500, "message": "Failed to add comment"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "Comment added successfully",
		"data": map[string]string{
			"messageId": messageID,
			"comment":   req.Comment,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
