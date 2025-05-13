package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getUserInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := rt.db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"code":    200,
		"message": "User info retrieved",
		"data": map[string]any{
			"userId":   user.ID,
			"username": user.Username,
			"name":     user.Name,
			"gender":   user.Gender,
			"photo":    user.Photo,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
