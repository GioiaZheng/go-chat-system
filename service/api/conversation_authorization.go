package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

var (
	ErrConversationUserRequired = errors.New("conversation membership requires user id")
	ErrConversationIDRequired   = errors.New("conversation membership requires conversation id")
	ErrConversationNotMember    = errors.New("user is not a conversation member")
)

// requireConversationMember verifies that userID belongs to conversationID.
// It is intentionally small so handlers can call it before reading or mutating
// conversation-scoped resources without duplicating membership checks.
func (rt *_router) requireConversationMember(userID, conversationID string) error {
	userID = strings.TrimSpace(userID)
	conversationID = strings.TrimSpace(conversationID)
	if userID == "" {
		return ErrConversationUserRequired
	}
	if conversationID == "" {
		return ErrConversationIDRequired
	}

	members, err := rt.db.GetConversationMembers(conversationID)
	if err != nil {
		return err
	}
	for _, memberID := range members {
		if strings.TrimSpace(memberID) == userID {
			return nil
		}
	}

	return ErrConversationNotMember
}

// loadAuthorizedMessage loads a message by ID and verifies the current user is
// a member of the message's conversation before the caller reads or mutates
// message-scoped resources such as comments.
func (rt *_router) loadAuthorizedMessage(w http.ResponseWriter, userID, messageID string) (models.Message, bool) {
	m, err := rt.db.GetMessageByID(strings.TrimSpace(messageID))
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "Message not found")
		return models.Message{}, false
	}
	if !rt.authorizeMessageConversation(w, userID, m) {
		return models.Message{}, false
	}
	return m, true
}

// authorizeMessageConversation verifies that the current user can access a
// previously-loaded message through its conversation membership.
func (rt *_router) authorizeMessageConversation(w http.ResponseWriter, userID string, m models.Message) bool {
	if strings.TrimSpace(userID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return false
	}
	if err := rt.requireConversationMember(userID, m.ConversationID); err != nil {
		if errors.Is(err, ErrConversationNotMember) {
			rt.sendError(w, http.StatusForbidden, "not your conversation")
			return false
		}
		rt.sendError(w, http.StatusInternalServerError, "Failed to verify conversation membership")
		return false
	}
	return true
}
