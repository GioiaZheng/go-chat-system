package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// CreateGroupRequest 定义创建群组时的请求体
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// createGroup 处理 POST /groups
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 解析请求体
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 如果没有提供群组成员，自动添加当前用户
	if len(req.Members) == 0 {
		req.Members = []string{userID}
	} else {
		// 确保当前用户在成员列表中
		found := false
		for _, id := range req.Members {
			if id == userID {
				found = true
				break
			}
		}
		if !found {
			req.Members = append(req.Members, userID)
		}
	}

	// 设置默认群组名称
	groupName := req.Name
	if groupName == "" {
		groupName = "Group created by " + userID
	}

	// 创建群组
	group := models.Group{
		Name: groupName,
	}
	err := rt.db.CreateGroup(group)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to create group"}`, http.StatusInternalServerError)
		return
	}

	// 获取刚创建的群组
	createdGroup, err := rt.db.GetGroupByName(groupName)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to retrieve created group"}`, http.StatusInternalServerError)
		return
	}

	// 添加成员到群组
	err = rt.db.AddGroupMembers(createdGroup.ID, req.Members)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to add members to group"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "Group created successfully",
		"data": map[string]interface{}{
			"groupId":   createdGroup.ID,
			"groupName": createdGroup.Name,
			"members":   req.Members,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
