package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// CreateGroupRequest 定义创建群组时的请求体
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// createGroup handles POST /groups/create
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 从Context中提取当前用户ID
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 解析请求体
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode create group request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 确保自己也在成员列表中
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

	// 群组名字，如果没提供，用默认
	groupName := req.Name
	if groupName == "" {
		groupName = "Group with " + userID
	}

	// Step1: 创建群组（插入groups表）
	group := models.Group{
		Name: groupName,
	}
	err := rt.db.CreateGroup(group)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to create group")
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Step2: 由于CreateGroup时group.ID没有被回写，需要重新查询
	createdGroup, err := rt.db.GetGroupByName(groupName)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to retrieve created group")
		http.Error(w, "Failed to retrieve created group", http.StatusInternalServerError)
		return
	}

	// Step3: 将所有成员加入到group_members表
	err = rt.db.AddGroupMembers(createdGroup.ID, req.Members)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to add group members")
		http.Error(w, "Failed to add members", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]any{
		"code":    201,
		"message": "Group created successfully",
		"data": map[string]string{
			"groupName": groupName,
		},
	}
	writeJSON(w, http.StatusCreated, resp)
}
