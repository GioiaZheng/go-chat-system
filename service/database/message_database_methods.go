package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// defaultPrivateLimit caps legacy direct-message history retrieval when callers
// rely on the compatibility helpers instead of paginated fetchers.
const defaultPrivateLimit = 50

// Message persistence helpers encapsulate insert, query, and ownership logic for
// both private and group conversations.

// Section: message send helpers that insert new private, conversation, or group rows.

// nullableReplyID converts a possibly-nil reply pointer into a driver-friendly value.
// database/sql does not accept *string directly, so we normalize to either a
// trimmed string or nil.
func nullableReplyID(reply *string) interface{} {
	if reply == nil {
		return nil
	}
	return NullIfEmpty(strings.TrimSpace(*reply))
}

// SendPrivateMessage stores a direct message between two users.
// It enforces that only sender_id and receiver_id are set, leaving group and conversation empty.
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	_, err := db.c.Exec(`
INSERT INTO messages (
id,
content,
file_url,
type,
status,
read,
sender_id,
                        receiver_id,
                        group_id,
                        conversation_id,
                        reply_to_id,
                        created_at
) VALUES (
?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, COALESCE(?, datetime('now'))
)
`,
		message.ID,
		message.Content,
		message.FileURL,
		message.Type,
		message.Status,
		message.Read,
		message.SenderID,
		message.ReceiverID,
		nullableReplyID(message.ReplyToID),
		NullIfEmpty(message.CreatedAt),
	)
	return err
}

// SendMessageToConversation persists a message inside a conversation thread.
// Group messages are funneled through this path as well so ordering remains consistent.
func (db *appdbimpl) SendMessageToConversation(message models.Message) error {
	_, err := db.c.Exec(`
INSERT INTO messages (
id,
content,
file_url,
type,
status,
read,
sender_id,
                        receiver_id,
                        group_id,
                        conversation_id,
                        reply_to_id,
                        created_at
) VALUES (
?, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, ?, COALESCE(?, datetime('now'))
)
`,
		message.ID,
		message.Content,
		message.FileURL,
		message.Type,
		message.Status,
		message.Read,
		message.SenderID,
		message.ConversationID,
		nullableReplyID(message.ReplyToID),
		NullIfEmpty(message.CreatedAt),
	)
	return err
}

// SendGroupMessage guards that a conversation has been selected before delegating
// to the shared SendMessageToConversation implementation.
func (db *appdbimpl) SendGroupMessage(message models.Message) error {
	if strings.TrimSpace(message.ConversationID) == "" {
		return errors.New("conversation_id required for group message")
	}
	return db.SendMessageToConversation(message)
}

// MarkConversationRead marks every message in the conversation as read for the caller
// (excluding the caller's own messages). It uses a coarse-grained update instead of
// per-message receipts to avoid real-time/stateful complexity.
func (db *appdbimpl) MarkConversationRead(conversationID, readerID string) error {
	conversationID = strings.TrimSpace(conversationID)
	readerID = strings.TrimSpace(readerID)
	if conversationID == "" || readerID == "" {
		return nil
	}

	_, err := db.c.Exec(`
UPDATE messages
   SET read = 1,
       status = 'read'
 WHERE conversation_id = ?
   AND sender_id <> ?
   AND read = 0
`, conversationID, readerID)
	return err
}

// Section: conversation reads with cursor support for paginating message history.

// GetMessagesByConversation fetches messages for a conversation using optional
// before/after cursors and a configurable limit. Results are ordered newest first.
func (db *appdbimpl) GetMessagesByConversation(
	convID, before, after string, limit int,
) ([]models.Message, error) {

	if limit <= 0 {
		limit = 20
	}
	convID = strings.TrimSpace(convID)
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)

	qb := `
SELECT
id,
content,
file_url,
type,
status,
read,
sender_id,
                        receiver_id,
                        group_id,
                        conversation_id,
                        reply_to_id,
                        created_at
        FROM messages
        WHERE conversation_id = ?`
	args := []interface{}{convID}

	if before != "" {
		qb += ` AND created_at < ?`
		args = append(args, before)
	} else if after != "" {
		qb += ` AND created_at > ?`
		args = append(args, after)
	}

	qb += ` ORDER BY created_at DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.c.Query(qb, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0, limit)
	for rows.Next() {
		var m models.Message
		var fileURL, recv, grp, conv, reply sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&fileURL,
			&m.Type,
			&m.Status,
			&m.Read,
			&m.SenderID,
			&recv,
			&grp,
			&conv,
			&reply,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fileURL.Valid {
			m.FileURL = fileURL.String
		}
		m.ReceiverID = recv.String
		m.GroupID = grp.String
		m.ConversationID = conv.String

		if fileURL.Valid {
			m.FileURL = fileURL.String
		}

		if reply.Valid {
			m.ReplyToID = &reply.String
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Section: legacy private and group readers retained for backward-compatible callers.

// GetPrivateConversation reads the recent direct-message history between two users.
func (db *appdbimpl) GetPrivateConversation(userID1, userID2 string) ([]models.Message, error) {
	return db.getPrivateConversationEx(context.Background(), userID1, userID2, defaultPrivateLimit, "")
}

// getPrivateConversationEx is the internal worker used by GetPrivateConversation;
// it accepts a context for future cancellation support and allows overriding the limit.
func (db *appdbimpl) getPrivateConversationEx(
	_ context.Context, userID1, userID2 string, limit int, _ string,
) ([]models.Message, error) {

	rows, err := db.c.Query(`
SELECT
id,
content,
file_url,
type,
status,
read,
sender_id,
            receiver_id,
            group_id,
            conversation_id,
            reply_to_id,
            created_at
        FROM messages
        WHERE (sender_id = ? AND receiver_id = ?)
           OR (sender_id = ? AND receiver_id = ?)
        ORDER BY created_at ASC, id ASC
        LIMIT ?
    `, userID1, userID2, userID2, userID1, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0, limit)
	for rows.Next() {
		var m models.Message
		var fileURL, reply sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&fileURL,
			&m.Type,
			&m.Status,
			&m.Read,
			&m.SenderID,
			&m.ReceiverID,
			&m.GroupID,
			&m.ConversationID,
			&reply,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fileURL.Valid {
			m.FileURL = fileURL.String
		}

		if reply.Valid {
			m.ReplyToID = &reply.String
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupConversation returns the ordered message history for a group conversation.
func (db *appdbimpl) GetGroupConversation(groupID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
        SELECT
            id,
            content,
file_url,
                        type,
                        status,
			read,
            sender_id,
            receiver_id,
            group_id,
            conversation_id,
            reply_to_id,
            created_at
        FROM messages
        WHERE group_id = ?
        ORDER BY created_at ASC, id ASC
    `, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0, 64)
	for rows.Next() {
		var m models.Message
		var fileURL, reply sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&fileURL,
			&m.Type,
			&m.Status,
			&m.Read,
			&m.SenderID,
			&m.ReceiverID,
			&m.GroupID,
			&m.ConversationID,
			&reply,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fileURL.Valid {
			m.FileURL = fileURL.String
		}

		if reply.Valid {
			m.ReplyToID = &reply.String
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Section: message lookup utilities.

func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
	var m models.Message
	var fileURL, recv, grp, conv, reply sql.NullString
	var readInt int

	err := db.c.QueryRow(`
SELECT
id,
content,
file_url,
type,
status,
read,
sender_id,
receiver_id,
group_id,
conversation_id,
reply_to_id,
created_at
FROM messages
WHERE id = ?
`, messageID).Scan(
		&m.ID,
		&m.Content,
		&fileURL,
		&m.Type,
		&m.Status,
		&readInt,
		&m.SenderID,
		&recv,
		&grp,
		&conv,
		&reply,
		&m.CreatedAt,
	)
	if err != nil {
		return models.Message{}, err
	}

	m.ReceiverID = recv.String
	m.GroupID = grp.String
	m.ConversationID = conv.String
	m.Read = readInt != 0

	if fileURL.Valid {
		m.FileURL = fileURL.String
	}

	if reply.Valid {
		m.ReplyToID = &reply.String
	}

	return m, nil
}

func (db *appdbimpl) GetAllMessages() ([]models.Message, error) {
	rows, err := db.c.Query(`
        SELECT
            id,
            content,
file_url,
                        type,
                        status,
            read,
			sender_id,
            receiver_id,
            group_id,
            conversation_id,
            reply_to_id,
            created_at
        FROM messages
        ORDER BY created_at DESC, id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0, 256)
	for rows.Next() {
		var m models.Message
		var fileURL, recv, grp, conv, reply sql.NullString
		var readInt int

		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&fileURL,
			&m.Type,
			&m.Status,
			&readInt,
			&m.SenderID,
			&recv,
			&grp,
			&conv,
			&reply,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}

		m.ReceiverID = recv.String
		m.GroupID = grp.String
		m.ConversationID = conv.String
		m.Read = readInt != 0

		if fileURL.Valid {
			m.FileURL = fileURL.String
		}

		if reply.Valid {
			m.ReplyToID = &reply.String
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Section: message forwarding plus ownership and deletion guards.

// ForwardMessage duplicates a message into another conversation, retaining content
// and metadata while assigning the current user as the sender.
func (db *appdbimpl) ForwardMessage(userID, messageID, toUserID, toGroupID string) error {
	orig, err := db.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	convID := strings.TrimSpace(toGroupID)
	if convID == "" {
		convID = strings.TrimSpace(orig.ConversationID)
	}
	if convID == "" {
		return errors.New("conversation_id required to forward")
	}

	_, err = db.c.Exec(`
INSERT INTO messages (
id,
content,
file_url,
type,
status,
read,
sender_id,
            receiver_id,
            group_id,
            conversation_id,
            reply_to_id,
            created_at
) VALUES (
lower(hex(randomblob(16))),
?, ?, ?, ?, ?, ?, NULL, NULL, ?, NULL, datetime('now')
)
`,
		orig.Content,
		orig.FileURL,
		orig.Type,
		orig.Status,
		0,
		userID,
		convID,
	)
	return err
}

// IsMessageOwner verifies whether the provided user created the target message.
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

// DeleteMessage removes a message if the caller is the original sender.
func (db *appdbimpl) DeleteMessage(userID, messageID string) error {
	ok, err := db.IsMessageOwner(userID, messageID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not the owner of the message")
	}
	_, err = db.c.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	return err
}

// Section: comment retrieval and mutation helpers.

// GetMessageComments returns the ordered list of comments attached to a message.
func (db *appdbimpl) GetMessageComments(messageID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
        SELECT
            id,
            type,
            content,
            sender_id,
            message_id,
            created_at
        FROM message_comments
        WHERE message_id = ?
        ORDER BY created_at ASC
    `, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0)
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(
			&m.ID,
			&m.Type,
			&m.Content,
			&m.SenderID,
			&m.ConversationID,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// CommentMessage creates a new comment on a message, preserving the provided type
// and content and assigning a server-generated timestamp.
func (db *appdbimpl) CommentMessage(
	messageID, userID, ctype, content string,
) error {
	_, err := db.c.Exec(`
        INSERT INTO message_comments (
            id,
            message_id,
            sender_id,
            type,
            content,
            created_at
        ) VALUES (
            lower(hex(randomblob(16))),
            ?, ?, ?, ?, datetime('now')
        )
    `, messageID, userID, ctype, content)
	return err
}

// UncommentMessage deletes the caller's comments linked to a specific message.
func (db *appdbimpl) UncommentMessage(messageID, userID string) error {
	messageID = strings.TrimSpace(messageID)
	userID = strings.TrimSpace(userID)
	if messageID == "" || userID == "" {
		return errors.New("message_id and user_id required to remove comments")
	}

	_, err := db.c.Exec(
		`DELETE FROM message_comments WHERE message_id = ? AND sender_id = ?`,
		messageID,
		userID,
	)
	return err
}
