package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type sendMessageRequest struct {
	// Official
	ConversationID string `json:"conversation_id,omitempty"`
	Content        string `json:"content,omitempty"`
	Type           string `json:"type,omitempty"` // text|image|video|file (default text)

	// Legacy aliases
	ChatType   string `json:"chat_type,omitempty"`
	TargetID   string `json:"target_id,omitempty"`
	ReceiverID string `json:"receiver_id,omitempty"`
	ToUserID   string `json:"to_user_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	Message    string `json:"message,omitempty"`
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req sendMessageRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	content := strings.TrimSpace(coalesce(req.Content, req.Message))
	if content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	msg := models.Message{
		ID:        uuid.NewString(),
		SenderID:  ctx.UserID,
		Content:   content,
		Type:      "text",
		Status:    "sent",
		CreatedAt: now,
	}

	// --- Official path: by conversation_id (DO NOT guess by prefix) ---
	if conv := strings.TrimSpace(req.ConversationID); conv != "" {
		msg.ConversationID = conv
		if err := rt.db.SendMessageToConversation(msg); err != nil {
			ctx.Logger.WithError(err).Error("failed to send message to conversation")
			rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
			return
		}
	} else {
		// --- Legacy path: chat_type + target_id / receiver_id / group_id ---
		chatType := strings.TrimSpace(strings.ToLower(req.ChatType))
		targetID := strings.TrimSpace(coalesce(req.TargetID, req.ReceiverID, req.ToUserID, req.GroupID))

		switch chatType {
		case "private", "direct", "dm":
			if targetID == "" {
				rt.sendError(w, http.StatusBadRequest, "target_id (or receiver_id/to_user_id) is required")
				return
			}
			msg.ReceiverID = targetID
			if err := rt.db.SendPrivateMessage(msg); err != nil {
				ctx.Logger.WithError(err).Error("failed to send private message")
				rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
				return
			}
		case "group", "grp":
			if targetID == "" {
				rt.sendError(w, http.StatusBadRequest, "group_id (or target_id) is required")
				return
			}
			// legacy: some callers may pass group_id; but DB should write via conversation_id
			msg.ConversationID = targetID
			if err := rt.db.SendMessageToConversation(msg); err != nil {
				ctx.Logger.WithError(err).Error("failed to send group(conversation) message")
				rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
				return
			}
		default:
			rt.sendError(w, http.StatusBadRequest, "conversation_id or valid chat_type required")
			return
		}
	}

	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Message sent successfully",
		"data": map[string]interface{}{
			"resource": msg,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode sendMessage response")
	}
}
