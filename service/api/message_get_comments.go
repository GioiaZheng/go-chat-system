package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type getCommentsResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// getMessageComments handles GET /messages/:id/comment
func (rt *_router) getMessageComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, _ reqcontext.RequestContext) {
	msgID := ps.ByName("id")
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "message id is required")
		return
	}

	// Validate message existence before returning comments
	if _, err := rt.db.GetMessageByID(msgID); err != nil {
		rt.sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// Placeholder payload (replace with real comments list when DB is ready)
	payload := map[string]interface{}{
		"message_id": msgID,
		"comments":   []interface{}{},
	}

	resp := getCommentsResponse{
		Code:    200,
		Message: "ok",
		Data:    payload,
	}

	// Use centralized writeJSON and check for error
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode message comments response")
	}
}
