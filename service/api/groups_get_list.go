package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getGroupsList handles GET /groups
func (rt *_router) getGroupsList(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 从数据库获取群组列表
	groups, err := rt.db.GetGroupsList(userID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to get groups list"}`, http.StatusInternalServerError)
		return
	}

	// 构建响应
	resp := map[string]interface{}{
		"code":    200,
		"message": "Groups list retrieved successfully",
		"data": map[string]interface{}{
			"groups": groups,
		},
	}

	// 设置响应头并返回JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
