package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SendMessageToGroup sends a group message (API exposed)
func (db *appdbimpl) SendMessageToGroup(message models.Message) error {
	_, err := db.c.Exec(
		`INSERT INTO messages (sender_id, group_id, content, created_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		message.SenderID, message.GroupID, message.Content,
	)
	return err
}
