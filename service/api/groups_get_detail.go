package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getGroupDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("id")

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get group details")
		return
	}

	// 只挑出需要的字段
	type SimpleUser struct {
		UserID    string `json:"userId"`
		UserName  string `json:"userName"`
		AvatarURL string `json:"avatarUrl"`
	}

	var members []SimpleUser
	for _, member := range group.Members {
		members = append(members, SimpleUser{
			UserID:    member.ID,
			UserName:  member.Username,
			AvatarURL: member.AvatarURL,
		})
	}

	// 构建简化版返回
	response := map[string]any{
		"code":    200,
		"message": "Group detail fetched successfully",
		"data": map[string]any{
			"id":      group.ID,
			"name":    group.Name,
			"members": members,
		},
	}

	writeJSON(w, http.StatusOK, response)
}
