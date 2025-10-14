// file: service/database/app_database.go
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	_ "github.com/mattn/go-sqlite3"
)

// AppDatabase is the contract used by the API layer.
// The concrete implementation is *appdbimpl (spread across multiple files in this package).
type AppDatabase interface {
	// lifecycle
	Close() error
	Ping() error

	// --- users ---
	GetUser(id string) (models.User, error) // implemented in user_database_methods.go
	GetUserByID(id string) (models.User, error)
	GetUserByCredentials(username, password string) (models.User, error)
	CreateUser(user models.User, password string) (models.User, error)
	UpdateUserName(userID, username string) error
	UpdateUserPhoto(userID, publicURL string) error
	CheckUserExists(username string) (bool, error)
	SearchUsers(ctx context.Context, me, query string) ([]models.User, error)
	GetUserIDFromIdentifier(identifier string) (string, error)

	// --- groups ---
	CreateGroup(group models.Group) error
	AddGroupMembers(groupID string, members []string) error
	GetGroup(id string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	GetGroupMembers(groupID string) ([]models.User, error)
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID, publicURL string) error
	LeaveGroup(groupID, userID string) error

	// --- conversations ---
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
	GetMyConversations(userID string) ([]models.Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)
	GetMessagesByConversation(conversationID, before, after string, limit int) ([]models.Message, error)

	// --- messages ---
	SendMessageToConversation(message models.Message) error
	SendPrivateMessage(message models.Message) error // legacy helper (may be no-op in your flow)
	SendGroupMessage(message models.Message) error   // shim to SendMessageToConversation

	GetAllMessages() ([]models.Message, error)
	GetPrivateConversation(userID1, userID2 string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)

	GetMessageByID(messageID string) (models.Message, error)
	GetMessageComments(messageID string) ([]models.Message, error)
	CommentMessage(messageID, userID, comment string) error
	UncommentMessage(messageID string) error
	ForwardMessage(userID, messageID, toUserID, toGroupID string) error

	IsMessageOwner(userID, messageID string) (bool, error)
	DeleteMessage(userID, messageID string) error
}

// Concrete implementation
type appdbimpl struct {
	c *sql.DB
}

// New wires a *sql.DB into our implementation, ensures PRAGMAs and schema exist.
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
			username   TEXT NOT NULL UNIQUE,
			password   TEXT,
			name       TEXT,
			avatar_url TEXT,
			photo      TEXT
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

		-- 新增：会话表
		CREATE TABLE IF NOT EXISTS conversations (
			id         TEXT PRIMARY KEY,
			name       TEXT,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- 新增：会话成员
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
			sender_id       TEXT NOT NULL,
			receiver_id     TEXT,
			group_id        TEXT,
			conversation_id TEXT,
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(sender_id)   REFERENCES users(id)   ON DELETE CASCADE,
			FOREIGN KEY(receiver_id) REFERENCES users(id)   ON DELETE SET NULL,
			FOREIGN KEY(group_id)    REFERENCES groups(id)  ON DELETE SET NULL
		);

		CREATE INDEX IF NOT EXISTS idx_messages_conv_created
			ON messages(conversation_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_messages_direct
			ON messages(sender_id, receiver_id, created_at DESC);
	`)
	return err
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

// Minimal helper used by API; implemented here for convenience.
func (db *appdbimpl) GetUserByID(userID string) (models.User, error) {
	var u models.User
	var name, avatarUrl, photo sql.NullString
	err := db.c.QueryRow(`
		SELECT id, username, name, avatar_url, photo
		  FROM users
		 WHERE id = ?
	`, userID).Scan(&u.ID, &u.Username, &name, &avatarUrl, &photo)
	if err != nil {
		return models.User{}, err
	}
	u.Name = name.String
	u.AvatarUrl = avatarUrl.String
	u.Photo = photo.String
	return u, nil
}
