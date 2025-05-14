package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// UserSearchRequest 表示搜索用户时的请求体
type UserSearchRequest struct {
	Query string `json:"query"` // 搜索查询
}

// UserSearchResponse 表示搜索结果的响应体
type UserSearchResponse struct {
	Users []struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		AvatarUrl string `json:"avatarUrl"`
	} `json:"users"`
}

// searchUsersHandler 处理 GET /users/search
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	// 解析请求体
	var req UserSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验查询内容
	if req.Query == "" {
		http.Error(w, `{"code": 400, "message": "Query parameter is required"}`, http.StatusBadRequest)
		return
	}

	// 从上下文获取 userID
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 搜索用户
	users, err := rt.db.SearchUsers(r.Context(), userID, req.Query) // 注意传递 context 和 userID
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to search users")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 构建响应
	resp := UserSearchResponse{
		Users: make([]struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			AvatarUrl string `json:"avatarUrl"`
		}, len(users)),
	}

	for i, user := range users {
		resp.Users[i] = struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			AvatarUrl string `json:"avatarUrl"`
		}{
			ID:        user.ID,
			Username:  user.Username,
			AvatarUrl: user.AvatarUrl,
		}
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response")
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}
