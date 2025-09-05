package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendMessageToUser is kept for compatibility with older call sites.
// REAL implementation (SendPrivateMessage) expects a models.Message and returns error only.
// We adapt the old triple-argument form by building the Message and delegating.
// NOTE: Prefer calling SendPrivateMessage directly in new code.
func (db *appdbimpl) SendMessageToUser(senderID, receiverID, content string) error {
	msg := models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	return db.SendPrivateMessage(msg)
}
