package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// sendMessageRequest supports multiple field aliases to be tolerant with different clients.
type sendMessageRequest struct {
	ChatType   string `json:"chat_type,omitempty"` // preferred
	Type       string `json:"type,omitempty"`      // alias
	TargetID   string `json:"target_id,omitempty"` // preferred
	ReceiverID string `json:"receiver_id,omitempty"`
	ToUserID   string `json:"to_user_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	Content    string `json:"content,omitempty"`   // preferred
	Message    string `json:"message,omitempty"`   // alias
}

// normalize returns (chatType, targetID, content)
func (p *sendMessageRequest) normalize() (string, string, string) {
	chatType := strings.TrimSpace(strings.ToLower(coalesce(p.ChatType, p.Type)))
	content := strings.TrimSpace(coalesce(p.Content, p.Message))

	// targetID precedence: target_id > receiver_id/to_user_id > group_id (and implies chatType=group)
	targetID := strings.TrimSpace(coalesce(p.TargetID, p.ReceiverID, p.ToUserID, p.GroupID))
	if chatType == "" && p.GroupID != "" {
		chatType = "group"
	}
	return chatType, targetID, content
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// sendMessage handles POST /messages.
// It accepts a tolerant JSON payload and delegates to DB layer.
// Response keeps the standard envelope: { code, message, data: { message } }.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req sendMessageRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	chatType, targetID, content := req.normalize()
	switch chatType {
	case "private", "direct", "dm":
		chatType = "private"
	case "group", "grp":
		chatType = "group"
	}

	if chatType != "private" && chatType != "group" {
		rt.sendError(w, http.StatusBadRequest, "chat_type must be 'private' or 'group'")
		return
	}
	if targetID == "" {
		rt.sendError(w, http.StatusBadRequest, "target_id (or receiver_id/group_id) is required")
		return
	}
	if content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	msg := models.Message{
		ID:       uuid.NewString(),
		SenderID: ctx.UserID,
		Content:  content,
	}
	var err error
	if chatType == "private" {
		msg.ReceiverID = targetID
		err = rt.db.SendPrivateMessage(msg)
	} else {
		msg.GroupID = targetID
		err = rt.db.SendGroupMessage(msg)
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to persist message")
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
		ctx.Logger.WithError(err).Error("failed to encode sendMessage response")
	}
}
