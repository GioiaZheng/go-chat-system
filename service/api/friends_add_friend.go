package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AddFriendRequest 是添加好友请求体
type AddFriendRequest struct {
	UserID string `json:"userId"` // 要添加的好友的用户ID
}

// addFriend handles POST /friends/add
func (rt *_router) addFriend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	currentUserID := GetUserIDFromContext(r.Context()) // 从 token/context 中获取当前登录用户ID
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode add friend request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		rt.baseLogger.Error("missing userId in add friend request")
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}

	// 不能加自己为好友
	if currentUserID == req.UserID {
		http.Error(w, "Cannot add yourself as friend", http.StatusBadRequest)
		return
	}

	// 调用数据库添加好友关系
	err := rt.db.AddFriend(currentUserID, req.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to add friend")
		http.Error(w, "Failed to add friend", http.StatusInternalServerError)
		return
	}

	// 成功返回
	resp := map[string]any{
		"code":    200,
		"message": "Friend added successfully",
		"data": map[string]string{
			"friendId": req.UserID,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}
