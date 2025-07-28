package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type SetUserNameRequest struct {
	Username string `json:"username"`
}

// PUT /users/set_username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SetUserNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验用户名合法性
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if req.Username == "" || !usernamePattern.MatchString(req.Username) || len(req.Username) > 50 {
		http.Error(w, `{"code": 400, "message": "Invalid username: must be 1-50 characters long and contain only letters, numbers, underscores, or hyphens"}`, http.StatusBadRequest)
		return
	}

	// 数据库更新
	if err := rt.db.UpdateUserName(userID, req.Username); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to update username")
		http.Error(w, `{"code": 500, "message": "Failed to update username"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Username updated successfully",
		"data": map[string]string{
			"username": req.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
