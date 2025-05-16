package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessageById handles GET /messages/:id
func (rt *_router) getMessageById(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("id")
	if messageID == "" {
		http.Error(w, `{"code": 400, "message": "Message ID is required"}`, http.StatusBadRequest)
		return
	}

	message, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get message by ID")
		http.Error(w, `{"code": 404, "message": "Message not found"}`, http.StatusNotFound)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"code":    200,
		"message": "Message fetched successfully",
		"data":    message,
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode message response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
