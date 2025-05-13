package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ForwardMessageRequest 转发请求体
type ForwardMessageRequest struct {
	ToUserID  string `json:"toUserId,omitempty"`
	ToGroupID string `json:"toGroupId,omitempty"`
}

// forwardMessage handles POST /messages/:id/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	var req ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode forward request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := rt.db.ForwardMessage(userID, messageID, req.ToUserID, req.ToGroupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to forward message")
		http.Error(w, "Failed to forward message", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"code":    200,
		"message": "Message forwarded successfully",
	}
	writeJSON(w, http.StatusOK, response)
}
