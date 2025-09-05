package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendMessageToGroup is kept for compatibility with older call sites.
// REAL implementation (SendGroupMessage) expects a models.Message and returns error only.
// We adapt the old triple-argument form by building the Message and delegating.
// NOTE: Prefer calling SendGroupMessage directly in new code.
func (db *appdbimpl) SendMessageToGroup(senderID, groupID, content string) error {
	msg := models.Message{
		SenderID: senderID,
		GroupID:  groupID,
		Content:  content,
	}
	return db.SendGroupMessage(msg)
}
