package database

import (
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SendGroupMessage is a backward-compatible shim.
// Internally delegates to SendMessageToConversation.
func (db *appdbimpl) SendGroupMessage(message models.Message) error {
	if message.ConversationID == "" {
		return fmt.Errorf("conversation_id required for group message")
	}
	return db.SendMessageToConversation(message)
}
