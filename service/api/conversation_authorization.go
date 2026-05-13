package api

import (
	"errors"
	"strings"
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
