package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getUserProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("userId")
	if userID == "" {
		http.Error(w, `{"code": 400, "message": "User ID is required"}`, http.StatusBadRequest)
		return
	}

	profile, err := rt.db.GetUserByID(userID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to get user profile"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "User profile retrieved successfully",
		"data": map[string]interface{}{
			"userId":    profile.ID,
			"username":  profile.Username,
			"name":      profile.Name,
			"email":     profile.Email,
			"gender":    profile.Gender,
			"avatarUrl": profile.AvatarUrl, // 注意字段名统一
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
