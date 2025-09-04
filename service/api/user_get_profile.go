package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getUserProfile handles GET /users/profile/:user_id
func (rt *_router) getUserProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("user_id")
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
			"user": map[string]interface{}{
				"id":        profile.ID,
				"username":  profile.Username,
				"name":      profile.Name,
				"email":     profile.Email,
				"gender":    profile.Gender,
				"avatarUrl": profile.AvatarUrl,
			},
		},
	}

	// Use writeJSON and handle error
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode user profile response")
	}
}
