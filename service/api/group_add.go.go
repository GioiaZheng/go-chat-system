package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// AddGroupMembersRequest 是添加群组成员的请求体
type AddGroupMembersRequest struct {
	UserIDs []string `json:"user_ids"`
}

// addToGroup 处理 POST /groups/:id/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		http.Error(w, `{"code": 400, "message": "Group ID is required"}`, http.StatusBadRequest)
		return
	}

	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 解析请求体
	var req AddGroupMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) == 0 {
		http.Error(w, `{"code": 400, "message": "User IDs cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// 尝试添加成员到群组
	err := rt.db.AddGroupMembers(groupID, req.UserIDs)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to add members to group"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "Members added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
