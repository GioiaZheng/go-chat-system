package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// AppDatabase defines the interface that all database implementations must satisfy.
// Every API handler only depends on this interface, not on the concrete implementation.
type AppDatabase interface {
	// --- Connection management ---
	Ping() error
	Close() error

	// --- Auth ---
	// CreateUser inserts a new user with a password.
	CreateUser(user models.User, password string) (models.User, error)
	// AuthenticateUser checks credentials and returns the user.
	AuthenticateUser(email, password string) (models.User, error)
	// GetUserIDFromIdentifier resolves a username/email to user ID.
	GetUserIDFromIdentifier(identifier string) (string, error)
	// CheckUserExists verifies if a user with the given username already exists.
	CheckUserExists(name string) (bool, error)
	// GetUserByCredentials fetches a user matching username+password.
	GetUserByCredentials(name, password string) (models.User, error)

	// --- Users ---
	UpdateUserName(userID, name string) error
	UpdateUserPhoto(userID, photoPath string) error
	GetUserByID(userID string) (models.User, error)
	GetUser(userID string) (models.User, error)

	// --- Groups ---
	CreateGroup(group models.Group) error
	AddMemberToGroup(groupID, userID, role string) error
	AddGroupMembers(groupID string, userIDs []string) error
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID, photoPath string) error
	GetGroup(groupID string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	LeaveGroup(groupID, userID string) error
	IsGroupMember(userID, groupID string) (bool, error)
	GetGroupByName(name string) (models.Group, error)
	// List members of a group.
	GetGroupMembers(groupID string) ([]models.User, error)
	// Backward-compatible alias.
	SetGroupPhoto(groupID, photoUrl string) error

	// --- Messages ---
	// Send (private/conversation)
	SendPrivateMessage(message models.Message) error
	SendMessageToConversation(message models.Message) error // unified conversation insert
	// Compatibility shim
	SendGroupMessage(message models.Message) error

	// Read
	GetPrivateConversation(userID1, userID2 string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)
	GetMessageByID(messageID string) (models.Message, error)
	GetMyConversations(userID string) ([]models.Conversation, error)

	// Comments / forward / ownership / delete
	CommentMessage(messageID, userID, comment string) error
	ForwardMessage(userID, messageID, toUserID, toGroupID string) error
	UncommentMessage(messageID string) error
	IsMessageOwner(userID, messageID string) (bool, error)
	DeleteMessage(userID, messageID string) error

	// Added: utility readers used by API handlers
	GetAllMessages() ([]models.Message, error)
	GetMessageComments(messageID string) ([]models.Message, error)

	// --- Conversations ---
	SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error)
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)
}
