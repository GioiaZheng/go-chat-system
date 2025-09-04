package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations handles GET /messages/conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get my conversations")
		http.Error(w, `{"code": 500, "message": "Failed to fetch conversations"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Conversations fetched successfully",
		"data": map[string]interface{}{
			"conversations": conversations,
		},
	}

	// Use writeJSON and handle error
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode conversations response")
	}
}
