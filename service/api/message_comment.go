package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// CommentMessageRequest 评论请求体
type CommentMessageRequest struct {
	Comment string `json:"comment"`
}

// commentMessage handles POST /messages/:id/comment
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("id")
	if messageID == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	var req CommentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode comment request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := rt.db.CommentMessage(messageID, userID, req.Comment)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to comment message")
		http.Error(w, "Failed to comment message", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"code":    200,
		"message": "Comment added successfully",
	}
	writeJSON(w, http.StatusOK, response)
}
