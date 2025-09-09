package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

const defaultPrivateLimit = 50

// GetPrivateConversation implements the AppDatabase interface.
// It adapts to the internal helper with sensible defaults.
func (db *appdbimpl) GetPrivateConversation(userID1 string, userID2 string) ([]models.Message, error) {
	return db.getPrivateConversationEx(context.Background(), userID1, userID2, defaultPrivateLimit, "")
}

// getPrivateConversationEx is the internal helper that supports paging and "before" cursor.
// It returns messages between userID1 and userID2 (both directions). Self-chat is naturally included.
func (db *appdbimpl) getPrivateConversationEx(ctx context.Context, userID1, userID2 string, limit int, before string) ([]models.Message, error) {
	args := []interface{}{userID1, userID2, userID2, userID1}
	q := `
		SELECT id, content, sender_id, receiver_id, conversation_id,
		       COALESCE(created_at, '') AS created_at,
		       type, status
		  FROM messages
		 WHERE (
		       (sender_id = ? AND receiver_id = ?)
		    OR (sender_id = ? AND receiver_id = ?)
		 )
	`
	if before != "" {
		q += " AND datetime(created_at) < datetime(?) "
		args = append(args, before)
	}
	q += " ORDER BY datetime(created_at) DESC, id DESC "
	if limit > 0 {
		q += " LIMIT ? "
		args = append(args, limit)
	}

	rows, err := db.c.QueryContext(ctx, q, args...)
	if err != nil {
		log.Println("GetPrivateConversation query error:", err)
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
