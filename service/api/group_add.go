package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Request body 兼容两种写法：{"user_ids": ["u1","u2"]} 或 {"user_id": "u1"}
type addMembersReq struct {
	UserIDs []string `json:"user_ids"`
	UserID  string   `json:"user_id"`
}

// POST /api/v1/groups/:id/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "group id is required")
		return
	}

	var req addMembersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// 兼容 user_ids / user_id 两种
	var userIDs []string
	if len(req.UserIDs) > 0 {
		userIDs = req.UserIDs
	} else if strings.TrimSpace(req.UserID) != "" {
		userIDs = []string{strings.TrimSpace(req.UserID)}
	}

	if len(userIDs) == 0 {
		rt.sendError(w, http.StatusBadRequest, "user_ids or user_id is required")
		return
	}

	// 调用数据库批量添加
	if err := rt.db.AddGroupMembers(groupID, userIDs); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to add members", err.Error())
		return
	}

	// 与 api.yaml 的 BaseSuccessResponse 对齐
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "Members added successfully",
	})
}
