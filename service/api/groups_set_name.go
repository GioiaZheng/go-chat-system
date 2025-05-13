package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// SetGroupNameRequest 设置群组名称请求体
type SetGroupNameRequest struct {
	Name string `json:"name"`
}

// setGroupName handles PUT /groups/:id/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	var req SetGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"code": 400, "message": "Group name is required"}`, http.StatusBadRequest)
		return
	}

	// 验证用户是否是群组成员或管理员
	isMember, err := rt.db.IsGroupMember(userID, groupID)
	if err != nil || !isMember {
		http.Error(w, `{"code": 403, "message": "You are not a member of this group"}`, http.StatusForbidden)
		return
	}

	// 更新群组名称
	err = rt.db.UpdateGroupName(groupID, req.Name)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to update group name"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Group name updated successfully",
	}

	writeJSON(w, http.StatusOK, resp)
}
