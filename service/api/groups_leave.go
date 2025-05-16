package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// leaveGroup handles DELETE /groups/:id/members
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 检查用户是否在群组中
	isMember, err := rt.db.IsGroupMember(userID, groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to check group membership")
		http.Error(w, `{"code": 500, "message": "Failed to check group membership"}`, http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, `{"code": 403, "message": "You are not a member of this group"}`, http.StatusForbidden)
		return
	}

	// 退出群组
	err = rt.db.LeaveGroup(groupID, userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to leave group")
		http.Error(w, `{"code": 500, "message": "Failed to leave group"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Left group successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
