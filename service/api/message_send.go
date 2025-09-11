package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type sendMessageBody struct {
	ConversationID string `json:"conversation_id"`
	TargetID       string `json:"target_id"` // legacy
	ChatType       string `json:"chat_type"` // legacy: private|group
	Content        string `json:"content"`
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

	// Trim all inputs
	body.ConversationID = strings.TrimSpace(body.ConversationID)
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.ChatType = strings.ToLower(strings.TrimSpace(body.ChatType))
	body.Content = strings.TrimSpace(body.Content)

	if body.Content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	var msg models.Message
	var err error

	switch {
	// 1) Canonical: send to conversation
	case body.ConversationID != "":
		msg, err = rt.db.SendMessageToConversation(r.Context(), userID, body.ConversationID, body.Content)

	// 2) Legacy: private chat
	case body.ChatType == "private" && body.TargetID != "":
		msg, err = rt.db.SendPrivateMessage(r.Context(), userID, body.TargetID, body.Content)

	// 3) Legacy: group chat
	case body.ChatType == "group" && body.TargetID != "":
		msg, err = rt.db.SendMessageToGroup(r.Context(), userID, body.TargetID, body.Content)

	default:
		rt.sendError(w, http.StatusBadRequest, "Specify either conversation_id, or (chat_type + target_id)")
		return
	}

	if err != nil {
		// Not found or membership errors should surface as 404 where possible
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
		"data": map[string]any{
			"message": msg,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode send message response")
	}
}
