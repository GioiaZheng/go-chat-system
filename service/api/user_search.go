package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// searchUsers 处理 GET /users/search
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 从查询参数获取搜索关键词
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid query parameters"}`, http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(queryParams.Get("q"))
	if query == "" {
		http.Error(w, `{"code": 400, "message": "Query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	// 校验搜索关键词
	queryPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-@.]+$`)
	if !queryPattern.MatchString(query) || len(query) > 255 {
		http.Error(w, `{"code": 400, "message": "Invalid query: must be 1-255 characters long and contain only letters, numbers, underscores, hyphens, @, or dots"}`, http.StatusBadRequest)
		return
	}

	// 执行用户搜索
	users, err := rt.db.SearchUsers(r.Context(), userID, query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to search users")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 构建响应
	var response []map[string]interface{}
	for _, user := range users {
		response = append(response, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"avatarUrl": user.AvatarUrl,
		})
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
