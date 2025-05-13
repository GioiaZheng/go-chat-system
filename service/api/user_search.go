package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

type searchUserRequest struct {
	Query string `json:"q"`
}

type searchedUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar_url,omitempty"`
}

func (rt *_router) searchUsersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	// 获取查询参数
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"code": 400, "message": "Invalid query parameter"}`, http.StatusBadRequest)
		return
	}

	// 验证用户身份
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 执行用户搜索
	results, err := rt.db.SearchUsers(r.Context(), query)

	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to search users"}`, http.StatusInternalServerError)
		return
	}

	// 构建响应
	var response []searchedUser
	for _, u := range results {
		response = append(response, searchedUser{
			ID:       u.ID,
			Username: u.Username,
			Name:     u.Name,
			Email:    u.Email,
			Avatar:   u.AvatarUrl,
		})
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Users retrieved successfully",
		"data":    response,
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SearchUsers searches for users based on a query and user ID
func (db *appdbimpl) SearchUsers(ctx context.Context, userID, query string) ([]models.User, error) {
	var users []models.User

	// 排除自己，防止自己出现在搜索结果中
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, gender
		FROM users
		WHERE (username LIKE ? OR name LIKE ? OR email LIKE ?)
		AND id != ?
	`, "%"+query+"%", "%"+query+"%", "%"+query+"%", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.AvatarUrl, &user.Gender); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
