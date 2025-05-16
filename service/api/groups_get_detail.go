package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupDetail 处理 GET /groups/:id
func (rt *_router) getGroupDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 检查用户是否是群组成员
	isMember, err := rt.db.IsGroupMember(userID, groupID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to check group membership"}`, http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, `{"code": 403, "message": "You are not a member of this group"}`, http.StatusForbidden)
		return
	}

	// 获取群组详情
	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to get group details"}`, http.StatusInternalServerError)
		return
	}

	// 构建成员信息
	members := make([]map[string]interface{}, 0)
	for _, member := range group.Members {
		members = append(members, map[string]interface{}{
			"userId":    member.UserID,
			"userName":  member.UserName,
			"avatarUrl": member.AvatarUrl,
		})
	}

	// 返回群组详情
	resp := map[string]interface{}{
		"code":    200,
		"message": "Group detail fetched successfully",
		"data": map[string]interface{}{
			"id":      group.ID,
			"name":    group.Name,
			"members": members,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
