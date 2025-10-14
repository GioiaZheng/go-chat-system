package database

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

const defaultPrivateLimit = 50

// nullIfEmpty is a small helper for nullable timestamps in INSERTs.
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

// SendPrivateMessage inserts a direct (1:1) message.
// NOTE: Our SQLite schema does NOT have type/status columns; we don't persist them.
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (
			id, content, sender_id, receiver_id, group_id, conversation_id, created_at
		) VALUES (
			?,  ?,       ?,         ?,           NULL,    NULL,            COALESCE(?, datetime('now'))
		)
	`, message.ID, message.Content, message.SenderID, message.ReceiverID, nullIfEmpty(message.CreatedAt))
	return err
}

// SendMessageToConversation inserts a message linked to a conversation.
// NOTE: No type/status persistence due to table schema.
func (db *appdbimpl) SendMessageToConversation(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (
			id, content, sender_id, receiver_id, group_id, conversation_id, created_at
		) VALUES (
			?,  ?,       ?,         NULL,        NULL,    ?,               COALESCE(?, datetime('now'))
		)
	`, message.ID, message.Content, message.SenderID, message.ConversationID, nullIfEmpty(message.CreatedAt))
	return err
}

// SendGroupMessage keeps backward compatibility; it delegates to SendMessageToConversation.
func (db *appdbimpl) SendGroupMessage(message models.Message) error {
	if message.ConversationID == "" {
		return errors.New("conversation_id required for group message")
	}
	return db.SendMessageToConversation(message)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   READ — BY CONVERSATION (used by GET /messages)
// ────────────────────────────────────────────────────────────────────────────────
//

// GetMessagesByConversation returns messages for a conversation with cursor pagination.
// - when before != "", return items with created_at < before (newer -> older)
// - when after  != "", return items with created_at > after  (older -> newer)
// If both are given, before takes precedence.
func (db *appdbimpl) GetMessagesByConversation(
	convID, before, after string, limit int,
) ([]models.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	convID = strings.TrimSpace(convID)
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)

	// Base WHERE
	qb := `
        SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
        FROM messages
        WHERE conversation_id = ?`
	args := []interface{}{convID}

	// Cursor predicates (exclusive)
	if before != "" {
		qb += ` AND created_at < ?`
		args = append(args, before)
	} else if after != "" {
		qb += ` AND created_at > ?`
		args = append(args, after)
	}

	// Stable order: created_at DESC, id DESC
	qb += ` ORDER BY created_at DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.c.Query(qb, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// receiver_id / group_id / conversation_id can be NULL -> scan via sql.NullString
	out := make([]models.Message, 0, limit)
	for rows.Next() {
		var m models.Message
		var recv, grp, conv sql.NullString
		if err := rows.Scan(
			&m.ID, &m.Content, &m.SenderID, &recv, &grp, &conv, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		m.ReceiverID = recv.String
		m.GroupID = grp.String
		m.ConversationID = conv.String
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   READ — PRIVATE / GROUP (legacy helpers; kept for completeness)
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) GetPrivateConversation(userID1, userID2 string) ([]models.Message, error) {
	return db.getPrivateConversationEx(context.Background(), userID1, userID2, defaultPrivateLimit, "")
}

func (db *appdbimpl) getPrivateConversationEx(
	_ context.Context, userID1, userID2 string, limit int, _ string,
) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
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
		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.SenderID,
			&m.ReceiverID,
			&m.GroupID,
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

func (db *appdbimpl) GetGroupConversation(groupID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
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
		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.SenderID,
			&m.ReceiverID,
			&m.GroupID,
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

// GetMessageByID with NULL-safe scanning (fixes 404 on NULL columns)
func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
	var m models.Message
	var recv, grp, conv sql.NullString

	err := db.c.QueryRow(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
		  FROM messages
		 WHERE id = ?
	`, messageID).Scan(
		&m.ID,
		&m.Content,
		&m.SenderID,
		&recv,
		&grp,
		&conv,
		&m.CreatedAt,
	)
	if err != nil {
		return models.Message{}, err
	}
	m.ReceiverID = recv.String
	m.GroupID = grp.String
	m.ConversationID = conv.String
	return m, nil
}

func (db *appdbimpl) GetAllMessages() ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
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
		var recv, grp, conv sql.NullString
		if err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.SenderID,
			&recv,
			&grp,
			&conv,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		m.ReceiverID = recv.String
		m.GroupID = grp.String
		m.ConversationID = conv.String
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

//
// ────────────────────────────────────────────────────────────────────────────────
/*  COMMENTS (lazy table creation)  */
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) ensureCommentsTable() error {
	_, err := db.c.Exec(`
		CREATE TABLE IF NOT EXISTS message_comments (
			id          TEXT PRIMARY KEY,
			message_id  TEXT NOT NULL,
			sender_id   TEXT NOT NULL,
			content     TEXT NOT NULL,
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(message_id) REFERENCES messages(id),
			FOREIGN KEY(sender_id)  REFERENCES users(id)
		)
	`)
	return err
}

func (db *appdbimpl) GetMessageComments(messageID string) ([]models.Message, error) {
	if err := db.ensureCommentsTable(); err != nil {
		return nil, err
	}

	rows, err := db.c.Query(`
		SELECT id, content, sender_id, created_at
		  FROM message_comments
		 WHERE message_id = ?
		 ORDER BY created_at ASC, id ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Message, 0, 16)
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (db *appdbimpl) CommentMessage(messageID, userID, comment string) error {
	if err := db.ensureCommentsTable(); err != nil {
		return err
	}
	_, err := db.c.Exec(`
		INSERT INTO message_comments (id, message_id, sender_id, content, created_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'))
	`, messageID, userID, comment)
	return err
}

func (db *appdbimpl) UncommentMessage(messageID string) error {
	if err := db.ensureCommentsTable(); err != nil {
		return err
	}
	_, err := db.c.Exec(`DELETE FROM message_comments WHERE message_id = ?`, messageID)
	return err
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   FORWARD / DELETE / OWNERSHIP
// ────────────────────────────────────────────────────────────────────────────────
//

// ForwardMessage now prefers toGroupID (target conversationId) and falls back to original.
func (db *appdbimpl) ForwardMessage(userID, messageID, toUserID, toGroupID string) error {
	orig, err := db.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// private forwarding
	if strings.TrimSpace(toUserID) != "" {
		_, err = db.c.Exec(`
			INSERT INTO messages (id, content, sender_id, receiver_id, group_id, conversation_id, created_at)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, NULL, NULL, datetime('now'))
		`, orig.Content, userID, toUserID)
		return err
	}

	// conversation forwarding
	convID := strings.TrimSpace(toGroupID)
	if convID == "" {
		convID = strings.TrimSpace(orig.ConversationID)
	}
	if convID == "" {
		return errors.New("conversation_id required to forward to group")
	}

	_, err = db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, receiver_id, group_id, conversation_id, created_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, NULL, NULL, ?, datetime('now'))
	`, orig.Content, userID, convID)
	return err
}

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

//
// ────────────────────────────────────────────────────────────────────────────────
//   CONVERSATIONS LIST (used by GET /conversations)
// ────────────────────────────────────────────────────────────────────────────────
//

// GetMyConversations returns conversation summaries for a user.
// 1) Try standard tables: conversations + conversation_members
// 2) Fallback: aggregate from messages when (1) not available
func (db *appdbimpl) GetMyConversations(userID string) ([]models.Conversation, error) {
	type row struct {
		id, name, avatar                          sql.NullString
		lastID, lastContent, lastSender, lastConv sql.NullString
		lastAt                                    sql.NullString
	}

	// Path 1: standard tables
	rows, err := db.c.Query(`
		SELECT
			c.id,
			c.name,
			c.avatar_url,
			(SELECT id         FROM messages m WHERE m.conversation_id = c.id ORDER BY created_at DESC, id DESC LIMIT 1) AS last_id,
			(SELECT content    FROM messages m WHERE m.conversation_id = c.id ORDER BY created_at DESC, id DESC LIMIT 1) AS last_content,
			(SELECT sender_id  FROM messages m WHERE m.conversation_id = c.id ORDER BY created_at DESC, id DESC LIMIT 1) AS last_sender,
			(SELECT conversation_id FROM messages m WHERE m.conversation_id = c.id ORDER BY created_at DESC, id DESC LIMIT 1) AS last_conv,
			(SELECT created_at FROM messages m WHERE m.conversation_id = c.id ORDER BY created_at DESC, id DESC LIMIT 1) AS last_at
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = ?
		ORDER BY c.name COLLATE NOCASE ASC
	`, userID)
	if err == nil {
		defer rows.Close()
		var out []models.Conversation
		for rows.Next() {
			var rr row
			if err := rows.Scan(&rr.id, &rr.name, &rr.avatar, &rr.lastID, &rr.lastContent, &rr.lastSender, &rr.lastConv, &rr.lastAt); err != nil {
				return nil, err
			}
			c := models.Conversation{
				ID:        strings.TrimSpace(rr.id.String),
				Name:      strings.TrimSpace(rr.name.String),
				AvatarUrl: strings.TrimSpace(rr.avatar.String),
			}
			if rr.lastID.Valid {
				c.LastMessage = &models.Message{
					ID:             strings.TrimSpace(rr.lastID.String),
					Content:        strings.TrimSpace(rr.lastContent.String),
					SenderID:       strings.TrimSpace(rr.lastSender.String),
					ConversationID: strings.TrimSpace(rr.lastConv.String),
					CreatedAt:      strings.TrimSpace(rr.lastAt.String),
				}
			}
			out = append(out, c)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	// Path 2: fallback — aggregate from messages table (uses membership filter)
	msgRows, err2 := db.c.Query(`
		SELECT conversation_id, id, content, sender_id, created_at
		  FROM messages
		 WHERE conversation_id IN (
			 SELECT DISTINCT conversation_id
			   FROM conversation_members
			  WHERE user_id = ?
		 )
		 ORDER BY created_at DESC, id DESC
	`, userID)
	if err2 != nil {
		// If both fail, return the first error (closer to root cause).
		return nil, err
	}
	defer msgRows.Close()

	type agg struct {
		name, avatar string
		last         *models.Message
	}
	m := make(map[string]*agg)

	for msgRows.Next() {
		var convID, mid, content, sender, createdAt sql.NullString
		if scanErr := msgRows.Scan(&convID, &mid, &content, &sender, &createdAt); scanErr != nil {
			return nil, scanErr
		}
		id := strings.TrimSpace(convID.String)
		if id == "" {
			continue
		}
		a := m[id]
		if a == nil {
			a = &agg{name: "Conversation", avatar: "", last: nil}
			m[id] = a
		}
		// first seen is the latest because of DESC order
		if a.last == nil {
			a.last = &models.Message{
				ID:             strings.TrimSpace(mid.String),
				Content:        strings.TrimSpace(content.String),
				SenderID:       strings.TrimSpace(sender.String),
				ConversationID: id,
				CreatedAt:      strings.TrimSpace(createdAt.String),
			}
		}
	}
	if err := msgRows.Err(); err != nil {
		return nil, err
	}

	out := make([]models.Conversation, 0, len(m))
	for id, a := range m {
		out = append(out, models.Conversation{
			ID:          id,
			Name:        a.name,
			AvatarUrl:   a.avatar,
			LastMessage: a.last,
		})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}
