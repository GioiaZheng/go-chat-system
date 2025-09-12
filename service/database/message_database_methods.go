package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

const defaultPrivateLimit = 50

// nullIfEmpty is a small helper used when inserting nullable timestamps.
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   SEND (INSERTS)
// ────────────────────────────────────────────────────────────────────────────────
//

// SendPrivateMessage inserts a direct message to a single receiver (legacy private chat).
// Interface: AppDatabase.SendPrivateMessage(message models.Message) error
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	if message.Type == "" {
		message.Type = "text"
	}
	if message.Status == "" {
		message.Status = "sent"
	}
	_, err := db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, receiver_id, type, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, datetime('now')))
	`, message.ID, message.Content, message.SenderID, message.ReceiverID,
		message.Type, message.Status, nullIfEmpty(message.CreatedAt))
	return err
}

// SendMessageToConversation inserts a message into a conversation (group/multi-party).
// Interface: AppDatabase.SendMessageToConversation(message models.Message) error
func (db *appdbimpl) SendMessageToConversation(message models.Message) error {
	if message.Type == "" {
		message.Type = "text"
	}
	if message.Status == "" {
		message.Status = "sent"
	}
	_, err := db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, conversation_id, type, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, datetime('now')))
	`, message.ID, message.Content, message.SenderID, message.ConversationID,
		message.Type, message.Status, nullIfEmpty(message.CreatedAt))
	return err
}

// SendGroupMessage is a backward-compatible shim that delegates to SendMessageToConversation.
// Interface: AppDatabase.SendGroupMessage(message models.Message) error
func (db *appdbimpl) SendGroupMessage(message models.Message) error {
	if message.ConversationID == "" {
		return fmt.Errorf("conversation_id required for group message")
	}
	return db.SendMessageToConversation(message)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   READ (QUERIES)
// ────────────────────────────────────────────────────────────────────────────────
//

// GetPrivateConversation returns ordered messages between two users (legacy private).
// Interface: AppDatabase.GetPrivateConversation(userID1, userID2 string) ([]models.Message, error)
func (db *appdbimpl) GetPrivateConversation(userID1 string, userID2 string) ([]models.Message, error) {
	return db.getPrivateConversationEx(context.Background(), userID1, userID2, defaultPrivateLimit, "")
}

// Internal helper with pagination token and limit (can be extended later).
func (db *appdbimpl) getPrivateConversationEx(
	_ context.Context, userID1, userID2 string, limit int, _ string,
) ([]models.Message, error) {

	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, conversation_id, created_at, type, status
		  FROM messages
		 WHERE (sender_id = ? AND receiver_id = ?)
		    OR (sender_id = ? AND receiver_id = ?)
		 ORDER BY created_at ASC
		 LIMIT ?
	`, userID1, userID2, userID2, userID1, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		var convID sql.NullString
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ReceiverID, &convID, &m.CreatedAt, &m.Type, &m.Status); err != nil {
			return nil, err
		}
		m.ConversationID = convID.String // may be empty for legacy private
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupConversation returns the ordered messages of a group.
// Interface: AppDatabase.GetGroupConversation(groupID string) ([]models.Message, error)
func (db *appdbimpl) GetGroupConversation(groupID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, group_id, conversation_id, created_at
		  FROM messages
		 WHERE group_id = ?
		 ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.GroupID, &m.ConversationID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMessageByID returns a single message by its id.
// Interface: AppDatabase.GetMessageByID(messageID string) (models.Message, error)
func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
	var m models.Message
	// Use sql.NullString to tolerate legacy rows missing some columns.
	var recv, grp, conv sql.NullString
	if err := db.c.QueryRow(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at, type, status
		  FROM messages
		 WHERE id = ?
	`, messageID).Scan(&m.ID, &m.Content, &m.SenderID, &recv, &grp, &conv, &m.CreatedAt, &m.Type, &m.Status); err != nil {
		return models.Message{}, err
	}
	m.ReceiverID = recv.String
	m.GroupID = grp.String
	m.ConversationID = conv.String
	return m, nil
}

// GetMyConversations returns conversation summaries for a user (very simple).
// Interface: AppDatabase.GetMyConversations(userID string) ([]models.Conversation, error)
func (db *appdbimpl) GetMyConversations(userID string) ([]models.Conversation, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.name, c.avatar_url,
		       COALESCE( (SELECT content
		                  FROM messages m
		                  WHERE m.conversation_id = c.id
		                  ORDER BY created_at DESC LIMIT 1), '' ) AS last_message
		  FROM conversations c
		  JOIN conversation_members cm ON cm.conversation_id = c.id
		 WHERE cm.user_id = ?
		 ORDER BY c.name COLLATE NOCASE ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.Name, &c.AvatarUrl, &c.LastMessage); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAllMessages returns all messages (for admin/tests).
// Not required by the interface, but useful and used by some tooling.
func (db *appdbimpl) GetAllMessages() ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
		  FROM messages
		 ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.ConversationID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMessageComments returns comments for a message if you keep them in separate table.
// (If your schema keeps comments inline, feel free to adjust or return an empty slice.)
func (db *appdbimpl) GetMessageComments(messageID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, conversation_id, created_at
		  FROM message_comments
		 WHERE message_id = ?
		 ORDER BY created_at ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ConversationID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   UPDATE / MISC
// ────────────────────────────────────────────────────────────────────────────────
//

// CommentMessage attaches a comment to a message (simple schema).
// Interface: AppDatabase.CommentMessage(messageID, userID, comment string) error
func (db *appdbimpl) CommentMessage(messageID, userID, comment string) error {
	_, err := db.c.Exec(`
		INSERT INTO message_comments (id, message_id, sender_id, content, created_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'))
	`, messageID, userID, comment)
	return err
}

// UncommentMessage deletes the comment record (simple version).
// Interface: AppDatabase.UncommentMessage(messageID string) error
func (db *appdbimpl) UncommentMessage(messageID string) error {
	_, err := db.c.Exec(`DELETE FROM message_comments WHERE message_id = ?`, messageID)
	return err
}

// ForwardMessage creates a copy of the original message to a new target.
// Interface: AppDatabase.ForwardMessage(userID, messageID, toUserID, toGroupID string) error
func (db *appdbimpl) ForwardMessage(userID, messageID, toUserID, toGroupID string) error {
	orig, err := db.GetMessageByID(messageID)
	if err != nil {
		return err
	}
	// Keep the same content/type; override sender and routing.
	if toUserID != "" {
		_, err = db.c.Exec(`
			INSERT INTO messages (id, content, sender_id, receiver_id, type, status, created_at)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, 'sent', datetime('now'))
		`, orig.Content, userID, toUserID, orig.Type)
		return err
	}
	if toGroupID != "" || orig.ConversationID != "" {
		convID := orig.ConversationID
		if convID == "" {
			// When forwarding a legacy message to a group conversation we need the conversation id.
			log.Printf("ForwardMessage: conversation_id empty for original message %s; forwarding to group requires a conv id", messageID)
			return errors.New("conversation_id required to forward to group")
		}
		_, err = db.c.Exec(`
			INSERT INTO messages (id, content, sender_id, conversation_id, type, status, created_at)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, 'sent', datetime('now'))
		`, orig.Content, userID, convID, orig.Type)
		return err
	}
	return errors.New("either toUserID or toGroupID must be provided")
}

// IsMessageOwner returns true if the message is authored by the given user.
// Interface: AppDatabase.IsMessageOwner(userID, messageID string) (bool, error)
func (db *appdbimpl) IsMessageOwner(userID, messageID string) (bool, error) {
	var cnt int
	if err := db.c.QueryRow(`
		SELECT COUNT(1)
		  FROM messages
		 WHERE id = ?
		   AND sender_id = ?
	`, messageID, userID).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// DeleteMessage removes a message (only owner can delete; owner check is done at API).
// Interface: AppDatabase.DeleteMessage(userID, messageID string) error
func (db *appdbimpl) DeleteMessage(userID, messageID string) error {
	// Optional safety: double-check ownership here too.
	owner, err := db.IsMessageOwner(userID, messageID)
	if err != nil {
		return err
	}
	if !owner {
		return errors.New("not the owner of the message")
	}
	_, err = db.c.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	return err
}

