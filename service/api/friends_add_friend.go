package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// AddFriendRequest 是添加好友请求体
type AddFriendRequest struct {
	UserID string `json:"userId"`
}

// addFriend 处理 POST /friends/add
func (rt *_router) addFriend(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 解析请求体
	var req AddFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.UserID == userID {
		http.Error(w, `{"code": 400, "message": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// 尝试添加好友
	err := rt.db.AddFriend(userID, req.UserID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to add friend"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    200,
		"message": "Friend added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
