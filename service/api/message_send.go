package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// sendMessage handles POST /messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ToUserID  string `json:"toUserId"`
		ToGroupID string `json:"toGroupId"`
		Content   string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode send request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	message := models.Message{
		SenderID:   userID,
		Content:    req.Content,
		ReceiverID: req.ToUserID,
		GroupID:    req.ToGroupID,
	}

	var err error
	if req.ToUserID != "" {
		err = rt.db.SendPrivateMessage(message)
	} else if req.ToGroupID != "" {
		err = rt.db.SendGroupMessage(message)
	} else {
		http.Error(w, "Must specify either toUserId or toGroupId", http.StatusBadRequest)
		return
	}

	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to send message")
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"code":    201,
		"message": "Message sent successfully",
	})
}
