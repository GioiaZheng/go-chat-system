package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// getMyConversations handles GET /messages/conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get my conversations")
		http.Error(w, "Failed to fetch conversations", http.StatusInternalServerError)
		return
	}

	response := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Conversations any `json:"conversations"`
		} `json:"data"`
	}{
		Code:    200,
		Message: "Conversations fetched successfully",
		Data: struct {
			Conversations any `json:"conversations"`
		}{
			Conversations: conversations,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode conversations response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
