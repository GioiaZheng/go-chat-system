package database

import (
    "context"
    "database/sql"
    "errors"
    "strings"

    "github.com/GioiaZheng/Wasa_proj/service/models"
)

const defaultPrivateLimit = 50

//
// ────────────────────────────────────────────────────────────────────────────────
//   SEND (INSERTS)
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
    _, err := db.c.Exec(`
        INSERT INTO messages (
            id, content, sender_id, receiver_id, group_id, conversation_id, created_at
        ) VALUES (
            ?, ?, ?, ?, NULL, NULL, COALESCE(?, datetime('now'))
        )
    `, message.ID, message.Content, message.SenderID, message.ReceiverID, NullIfEmpty(message.CreatedAt))
    return err
}

func (db *appdbimpl) SendMessageToConversation(message models.Message) error {
    _, err := db.c.Exec(`
        INSERT INTO messages (
            id, content, sender_id, receiver_id, group_id, conversation_id, created_at
        ) VALUES (
            ?, ?, ?, NULL, NULL, ?, COALESCE(?, datetime('now'))
        )
    `, message.ID, message.Content, message.SenderID, message.ConversationID, NullIfEmpty(message.CreatedAt))
    return err
}

func (db *appdbimpl) SendGroupMessage(message models.Message) error {
    if strings.TrimSpace(message.ConversationID) == "" {
        return errors.New("conversation_id required for group message")
    }
    return db.SendMessageToConversation(message)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   READ — BY CONVERSATION
// ────────────────────────────────────────────────────────────────────────────────
//

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
        SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
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
        var recv, grp, conv sql.NullString

        if err := rows.Scan(
            &m.ID, &m.Content, &m.SenderID,
            &recv, &grp, &conv, &m.CreatedAt,
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
//   PRIVATE & GROUP (legacy)
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
            &m.ID, &m.Content, &m.SenderID,
            &m.ReceiverID, &m.GroupID,
            &m.ConversationID, &m.CreatedAt,
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
            &m.ID, &m.Content, &m.SenderID,
            &m.ReceiverID, &m.GroupID,
            &m.ConversationID, &m.CreatedAt,
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

//
// ────────────────────────────────────────────────────────────────────────────────
//   GET MESSAGE BY ID
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
    var m models.Message
    var recv, grp, conv sql.NullString

    err := db.c.QueryRow(`
        SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
          FROM messages
         WHERE id = ?
    `, messageID).Scan(
        &m.ID, &m.Content, &m.SenderID,
        &recv, &grp, &conv, &m.CreatedAt,
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
            &m.ID, &m.Content, &m.SenderID,
            &recv, &grp, &conv, &m.CreatedAt,
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
//   COMMENTS (OpenAPI)
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) ensureCommentsTable() error {
    _, err := db.c.Exec(`
        CREATE TABLE IF NOT EXISTS message_comments (
            id          TEXT PRIMARY KEY,
            message_id  TEXT NOT NULL,
            sender_id   TEXT NOT NULL,
            type        TEXT NOT NULL,           -- "text" | "emoji"
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
        SELECT id, type, content, sender_id, created_at
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
        if err := rows.Scan(&m.ID, &m.Type, &m.Content, &m.SenderID, &m.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, m)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }
    return out, nil
}

func (db *appdbimpl) CommentMessage(messageID, userID, ctype, content string) error {
    if err := db.ensureCommentsTable(); err != nil {
        return err
    }

    _, err := db.c.Exec(`
        INSERT INTO message_comments (id, message_id, sender_id, type, content, created_at)
        VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, datetime('now'))
    `, messageID, userID, ctype, content)
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
//   FORWARD / DELETE
// ────────────────────────────────────────────────────────────────────────────────
//

func (db *appdbimpl) ForwardMessage(userID, messageID, toUserID, toGroupID string) error {
    orig, err := db.GetMessageByID(messageID)
    if err != nil {
        return err
    }

    if strings.TrimSpace(toUserID) != "" {
        _, err = db.c.Exec(`
            INSERT INTO messages (
                id, content, sender_id, receiver_id, group_id, conversation_id, created_at
            ) VALUES (
                lower(hex(randomblob(16))), ?, ?, ?, NULL, NULL, datetime('now')
            )
        `, orig.Content, userID, toUserID)
        return err
    }

    convID := strings.TrimSpace(toGroupID)
    if convID == "" {
        convID = strings.TrimSpace(orig.ConversationID)
    }
    if convID == "" {
        return errors.New("conversation_id required to forward to group")
    }

    _, err = db.c.Exec(`
        INSERT INTO messages (
            id, content, sender_id, receiver_id, group_id, conversation_id, created_at
        ) VALUES (
            lower(hex(randomblob(16))), ?, ?, NULL, NULL, ?, datetime('now')
        )
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
