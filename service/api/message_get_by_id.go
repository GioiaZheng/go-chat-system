package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// getMessageById handles GET /messages/:id
func (rt *_router) getMessageById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("id") // 统一使用id
	if messageID == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	message, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get message by ID")
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	response := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}{
		Code:    200,
		Message: "Message fetched successfully",
		Data:    message,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode message response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
