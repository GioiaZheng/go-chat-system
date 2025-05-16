package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// UpdateGroupNameRequest 更新群组名称请求体
type UpdateGroupNameRequest struct {
	Name string `json:"name"`
}

// setGroupName 处理 PUT /groups/:groupId/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, `{"code": 400, "message": "Missing group ID"}`, http.StatusBadRequest)
		return
	}

	// 校验 groupId 格式
	groupIDPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !groupIDPattern.MatchString(groupID) || len(groupID) > 64 {
		http.Error(w, `{"code": 400, "message": "Invalid group ID format"}`, http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req UpdateGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验群组名称
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9\s_-]+$`)
	if req.Name == "" || !namePattern.MatchString(req.Name) || len(req.Name) > 100 {
		http.Error(w, `{"code": 400, "message": "Invalid name: must be 1-100 characters long and match the pattern [a-zA-Z0-9\\s_-]+"}`, http.StatusBadRequest)
		return
	}

	// 更新群组名称
	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to update group name")
		http.Error(w, `{"code": 500, "message": "Failed to update group name"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    200,
		"message": "Group name updated successfully",
		"data": map[string]string{
			"name": req.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
