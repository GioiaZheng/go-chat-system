package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getUserInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := rt.db.GetUser(userID)
	if err != nil {
		http.Error(w, `{"code":500,"message":"Failed to get user info"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "User info retrieved",
		"data": map[string]interface{}{
			"userId":    user.ID,
			"username":  user.Username,
			"name":      user.Name,
			"gender":    user.Gender,
			"photo":     user.AvatarUrl, // 用 AvatarUrl 作为 photo
			"email":     user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
