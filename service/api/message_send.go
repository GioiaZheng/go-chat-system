package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type sendMessageBody struct {
	ConversationID string `json:"conversation_id"`
	TargetID       string `json:"target_id"` // legacy
	ChatType       string `json:"chat_type"` // legacy: private|group
	Content        string `json:"content"`
	Type           string `json:"type,omitempty"`   // optional, defaults to "text"
	Status         string `json:"status,omitempty"` // optional, defaults to "sent"
}

func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	var body sendMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Normalize inputs
	body.ConversationID = strings.TrimSpace(body.ConversationID)
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.ChatType = strings.ToLower(strings.TrimSpace(body.ChatType))
	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	msg := models.Message{
		ID:             uuid.NewString(),
		Content:        body.Content,
		SenderID:       userID,
		ConversationID: body.ConversationID,
		ReceiverID:     "", // set below if needed
		Type:           strings.TrimSpace(body.Type),
		Status:         strings.TrimSpace(body.Status),
		// CreatedAt:   optional; DB COALESCE will default to now()
	}

	var err error
	switch {
	// 1) Canonical: conversation message
	case msg.ConversationID != "":
		err = rt.db.SendMessageToConversation(msg)

	// 2) Legacy: private (chat_type=private + target_id)
	case body.ChatType == "private" && body.TargetID != "":
		msg.ReceiverID = body.TargetID
		err = rt.db.SendPrivateMessage(msg)

	// 3) Legacy: group (chat_type=group + target_id)
	case body.ChatType == "group" && body.TargetID != "":
		// For backward compatibility we map "group target_id" as conversation_id.
		msg.ConversationID = body.TargetID
		err = rt.db.SendGroupMessage(msg)

	default:
		rt.sendError(w, http.StatusBadRequest, "Specify either conversation_id, or (chat_type + target_id)")
		return
	}

	if err != nil {
		// Try to classify not-found
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendError(w, http.StatusNotFound, "Conversation or target not found")
			return
		}
		ctx.Logger.WithError(err).Error("failed to send message")
		rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Message sent",
		"data": map[string]interface{}{
			"message": msg,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode send message response")
	}
}
