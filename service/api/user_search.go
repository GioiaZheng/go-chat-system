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

// DataResponse is the standard API response format
type DataResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

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

	queryPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-@.]+$`)
	if !queryPattern.MatchString(query) || len(query) > 255 {
		http.Error(w, `{"code": 400, "message": "Invalid query"}`, http.StatusBadRequest)
		return
	}

	users, err := rt.db.SearchUsers(r.Context(), userID, query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to search users")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var response []map[string]interface{}
	for _, user := range users {
		response = append(response, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"gender":    user.Gender,
			"avatarUrl": user.AvatarUrl,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    200,
		"message": "Search successful",
		"items":   response,
	})

}
