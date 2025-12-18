package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	_ "github.com/mattn/go-sqlite3"
)

// AppDatabase defines the database operations required by the API layer.
type AppDatabase interface {
	// lifecycle
	Close() error
	Ping() error

	// Section: users
	GetUser(id string) (models.User, error)
	GetUserByID(id string) (models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUserName(userID, name string) error
	UpdateUserPhoto(userID, publicURL string) error
	CheckUserExists(name string) (bool, error)
	SearchUsers(ctx context.Context, me, query string) ([]models.User, error)
	GetUserIDFromIdentifier(identifier string) (string, error)

	// Section: groups
	CreateGroup(group models.Group) error
	AddGroupMembers(groupID string, members []string) error
	GetGroup(id string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	GetGroupMembers(groupID string) ([]models.User, error)
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID, publicURL string) error
	LeaveGroup(groupID, userID string) error
	IsGroupMember(userID, groupID string) (bool, error)

	// Section: conversations
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
	GetMyConversations(userID string) ([]models.Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)
	GetMessagesByConversation(conversationID, before, after string, limit int) ([]models.Message, error)
	DeleteConversation(conversationID string) error

	// Section: messages
	MarkConversationRead(conversationID, readerID string) error
	SendMessageToConversation(message models.Message) error
	SendPrivateMessage(message models.Message) error
	SendGroupMessage(message models.Message) error

	GetAllMessages() ([]models.Message, error)
	GetPrivateConversation(userID1, userID2 string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)

	GetMessageByID(messageID string) (models.Message, error)
	GetMessageComments(messageID string) ([]models.Message, error)
	CommentMessage(messageID, userID, ctype, content string) error
	UncommentMessage(messageID string) error
	ForwardMessage(userID, messageID, toUserID, toGroupID string) error

	IsMessageOwner(userID, messageID string) (bool, error)
	DeleteMessage(userID, messageID string) error
}

// appdbimpl is the concrete implementation backed by an SQLite connection.
type appdbimpl struct {
	c *sql.DB
}

// New wires a *sql.DB into the implementation, enabling PRAGMAs and ensuring
// that the schema is present before returning a ready-to-use database handle.
func New(db *sql.DB) (*appdbimpl, error) {
	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Enable FK constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Ensure schema
	if err := initializeTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return &appdbimpl{c: db}, nil
}

func initializeTables(db *sql.DB) error {
	_, err := db.Exec(`
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL UNIQUE,
			avatar_url TEXT
		);

		CREATE TABLE IF NOT EXISTS groups (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS group_members (
			group_id TEXT,
			user_id  TEXT,
			role     TEXT,
			PRIMARY KEY (group_id, user_id),
			FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id)  REFERENCES users(id)  ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS conversations (
			id         TEXT PRIMARY KEY,
			name       TEXT,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS conversation_members (
			conversation_id TEXT NOT NULL,
			user_id         TEXT NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id)         REFERENCES users(id)         ON DELETE CASCADE
		);

        CREATE TABLE IF NOT EXISTS messages (
            id              TEXT PRIMARY KEY,
            content         TEXT NOT NULL,
            file_url        TEXT,
            type            TEXT DEFAULT 'text',
            status          TEXT DEFAULT 'sent',
            read            INTEGER DEFAULT 0,
            sender_id       TEXT NOT NULL,
            receiver_id     TEXT,
            group_id        TEXT,
			conversation_id TEXT,
			reply_to_id     TEXT,
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY(sender_id)       REFERENCES users(id)         ON DELETE CASCADE,
			FOREIGN KEY(receiver_id)     REFERENCES users(id)         ON DELETE SET NULL,
			FOREIGN KEY(group_id)        REFERENCES groups(id)        ON DELETE SET NULL,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_messages_conv_created
			ON messages(conversation_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_messages_direct
			ON messages(sender_id, receiver_id, created_at DESC);

		CREATE TABLE IF NOT EXISTS message_comments (
			id          TEXT PRIMARY KEY,
			message_id  TEXT NOT NULL,
			sender_id   TEXT NOT NULL,
			type        TEXT NOT NULL,
			content     TEXT NOT NULL,
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY(sender_id)  REFERENCES users(id)    ON DELETE CASCADE
		);

	`)
	if err != nil {
		return err
	}

	if err := ensureMessageReplyColumn(db); err != nil {
		return err
	}

	if err := ensureMessageFileURLColumn(db); err != nil {
		return err
	}

	if err := ensureGroupConversationColumn(db); err != nil {
		return err
	}

	if err := ensureMessageTypeColumn(db); err != nil {
		return err
	}

	if err := ensureMessageStatusColumn(db); err != nil {
		return err
	}

	if err := ensureMessageReadColumn(db); err != nil {
		return err
	}

	return nil
}

func ensureMessageTypeColumn(db *sql.DB) error {
	exists, err := columnExists(db, "messages", "type")
	if err != nil {
		return fmt.Errorf("failed to inspect messages table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN type TEXT DEFAULT 'text';`); err != nil {
		return fmt.Errorf("failed to add type column: %w", err)
	}
	return nil
}

func ensureMessageStatusColumn(db *sql.DB) error {
	exists, err := columnExists(db, "messages", "status")
	if err != nil {
		return fmt.Errorf("failed to inspect messages table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN status TEXT DEFAULT 'sent';`); err != nil {
		return fmt.Errorf("failed to add status column: %w", err)
	}

	return nil
}

func ensureMessageReadColumn(db *sql.DB) error {
	exists, err := columnExists(db, "messages", "read")
	if err != nil {
		return fmt.Errorf("failed to inspect messages table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN read INTEGER DEFAULT 0;`); err != nil {
		return fmt.Errorf("failed to add read column: %w", err)
	}

	return nil
}

func ensureMessageReplyColumn(db *sql.DB) error {
	exists, err := columnExists(db, "messages", "reply_to_id")
	if err != nil {
		return fmt.Errorf("failed to inspect messages table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN reply_to_id TEXT;`); err != nil {
		return fmt.Errorf("failed to add reply_to_id column: %w", err)
	}

	return nil
}

func ensureMessageFileURLColumn(db *sql.DB) error {
	exists, err := columnExists(db, "messages", "file_url")
	if err != nil {
		return fmt.Errorf("failed to inspect messages table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN file_url TEXT;`); err != nil {
		return fmt.Errorf("failed to add file_url column: %w", err)
	}

	return nil
}

func ensureGroupConversationColumn(db *sql.DB) error {
	exists, err := columnExists(db, "groups", "conversation_id")
	if err != nil {
		return fmt.Errorf("failed to inspect groups table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE groups ADD COLUMN conversation_id TEXT;`); err != nil {
		return fmt.Errorf("failed to add conversation_id to groups: %w", err)
	}

	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// lifecycle
func (db *appdbimpl) Close() error { return db.c.Close() }
func (db *appdbimpl) Ping() error  { return db.c.Ping() }

// helper reused by insert functions
func NullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (db *appdbimpl) InitDefaultUsers() error {
	// count users
	var count int
	row := db.c.QueryRow("SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already have users, skip
	}

	// default demo users
	defaultUsers := []struct {
		Username string
		Name     string
	}{
		{"alice", "Alice"},
		{"bob", "Bob"},
		{"charlie", "Charlie"},
	}

	for _, u := range defaultUsers {
		_, err := db.c.Exec(`INSERT INTO users (username, name, password) VALUES (?, ?, ?)`,
			u.Username, u.Name, "demo123")
		if err != nil {
			return err
		}
	}
	return nil
}
