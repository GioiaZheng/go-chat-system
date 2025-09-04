package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type SendMessageRequest struct {
	Content   string `json:"content"`
	ToUserID  string `json:"toUserId,omitempty"`
	ToGroupID string `json:"toGroupId,omitempty"`
}

// sendMessage handles POST /messages
// It sends a message either to a user (private) or to a group depending on payload.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendMessageRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"code":400,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, `{"code":400,"message":"Content is required"}`, http.StatusBadRequest)
		return
	}
	// Exactly one target must be provided
	if (req.ToUserID == "" && req.ToGroupID == "") || (req.ToUserID != "" && req.ToGroupID != "") {
		http.Error(w, `{"code":400,"message":"Provide either toUserId or toGroupId"}`, http.StatusBadRequest)
		return
	}

	// Dispatch based on target using AppDatabase interface methods
	if req.ToUserID != "" {
		msg := models.Message{
			SenderID:   ctx.UserID,
			ReceiverID: req.ToUserID,
			Content:    req.Content,
		}
		// Interface method name: SendPrivateMessage
		if err := rt.db.SendPrivateMessage(msg); err != nil {
			rt.baseLogger.WithError(err).Error("failed to send private message")
			http.Error(w, `{"code":500,"message":"Failed to send message"}`, http.StatusInternalServerError)
			return
		}
	} else {
		msg := models.Message{
			SenderID: ctx.UserID,
			GroupID:  req.ToGroupID,
			Content:  req.Content,
		}
		// Interface method name: SendGroupMessage
		if err := rt.db.SendGroupMessage(msg); err != nil {
			rt.baseLogger.WithError(err).Error("failed to send group message")
			http.Error(w, `{"code":500,"message":"Failed to send message"}`, http.StatusInternalServerError)
			return
		}
	}

	resp := map[string]interface{}{
		"code":    201,
		"message": "Message sent",
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode send message response")
	}
}
