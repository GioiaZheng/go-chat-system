package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type SetPhotoRequest struct {
	Photo string `json:"photo"`
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SetPhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Photo == "" {
		http.Error(w, "Photo is required", http.StatusBadRequest)
		return
	}

	err := rt.db.UpdateUserPhoto(userID, req.Photo)
	if err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"code":    200,
		"message": "Photo updated successfully",
		"data": map[string]string{
			"photo": req.Photo,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
