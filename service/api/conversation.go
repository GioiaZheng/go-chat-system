package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// StartConversationRequest 会话创建请求体
type StartConversationRequest struct {
	MemberIds []string `json:"memberIds"`
	Name      string   `json:"name"`
}

// startConversation 处理 POST /start_conversation
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 解析请求体
	var req StartConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验会话名称
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9\s'-]+$`)
	if req.Name == "" || !namePattern.MatchString(req.Name) || len(req.Name) > 100 {
		http.Error(w, `{"code": 400, "message": "Invalid name: must be 1-100 characters long and match the pattern [a-zA-Z0-9\\s'-]+"}`, http.StatusBadRequest)
		return
	}

	// 校验成员列表
	if len(req.MemberIds) == 0 || len(req.MemberIds) > 100 {
		http.Error(w, `{"code": 400, "message": "MemberIds must contain 1 to 100 user IDs"}`, http.StatusBadRequest)
		return
	}

	// 确保每个成员 ID 符合格式
	idPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	for _, memberID := range req.MemberIds {
		if !idPattern.MatchString(memberID) || len(memberID) > 64 {
			http.Error(w, `{"code": 400, "message": "Invalid member ID format"}`, http.StatusBadRequest)
			return
		}
	}

	// 创建会话
	conversation, err := rt.db.StartConversation(r.Context(), userID, req.MemberIds, req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to start conversation")
		http.Error(w, `{"code": 500, "message": "Failed to start conversation"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "Conversation started successfully",
		"data":    conversation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
