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

// sendMessageRequest supports both official OpenAPI fields and legacy aliases.
// Official: conversation_id, content, type
// Legacy:   chat_type + target_id/receiver_id/to_user_id/group_id + message
type sendMessageRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Content        string `json:"content,omitempty"`
	Type           string `json:"type,omitempty"` // text|image|video|file (default text)

	ChatType   string `json:"chat_type,omitempty"`
	TargetID   string `json:"target_id,omitempty"`
	ReceiverID string `json:"receiver_id,omitempty"`
	ToUserID   string `json:"to_user_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	Message    string `json:"message,omitempty"`
}

// coalesce picks the first non-empty trimmed string from the list.
func coalesce(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// sendMessage handles POST /messages.
// Flow:
//  1. Ensure the request is authenticated (ctx.UserID).
//  2. Parse and normalize the JSON body.
//  3. Choose official path (conversation_id) or legacy path (chat_type + target).
//  4. Persist the message using db.* helpers.
//  5. Return 201 Created with the message resource.
func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Auth check
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2) Parse request
	var req sendMessageRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	content := strings.TrimSpace(coalesce(req.Content, req.Message))
	if content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Prepare message object
	now := time.Now().UTC().Format(time.RFC3339)
	msg := models.Message{
		ID:        uuid.NewString(),
		SenderID:  ctx.UserID,
		Content:   content,
		Type:      "text",
		Status:    "sent",
		CreatedAt: now,
	}

	// 3) Official path: use conversation_id
	if conv := strings.TrimSpace(req.ConversationID); conv != "" {
		msg.ConversationID = conv
		if err := rt.db.SendMessageToConversation(msg); err != nil {
			ctx.Logger.WithError(err).Error("failed to send message to conversation")
			rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
			return
		}
	} else {
		// 4) Legacy path: use chat_type + target
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
			msg.ConversationID = targetID
			if err := rt.db.SendMessageToConversation(msg); err != nil {
				ctx.Logger.WithError(err).Error("failed to send group message")
				rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
				return
			}
		default:
			rt.sendError(w, http.StatusBadRequest, "conversation_id or valid chat_type required")
			return
		}
	}

	// 5) Success response
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
