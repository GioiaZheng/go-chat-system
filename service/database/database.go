package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// AppDatabase interface
// AppDatabase interface
type AppDatabase interface {
	Ping() error
	Close() error

	// Auth
	CreateUser(user models.User, password string) (models.User, error)
	AuthenticateUser(email, password string) (models.User, error)
	GetUserIDFromIdentifier(identifier string) (string, error)
	CheckUserExists(name string) (bool, error)
	GetUserByCredentials(name, password string) (models.User, error)

	// Users
	UpdateUserName(userID, name string) error
	UpdateUserPhoto(userID, photoPath string) error
	GetUserByID(userID string) (models.User, error)
	GetUser(userID string) (models.User, error)

	// Groups
	CreateGroup(group models.Group) error
	AddMemberToGroup(groupID, userID, role string) error
	AddGroupMembers(groupID string, userIDs []string) error
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID, photoPath string) error
	GetGroup(groupID string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	LeaveGroup(groupID, userID string) error
	IsGroupMember(userID, groupID string) (bool, error)

	// Messages
	SendPrivateMessage(message models.Message) error
	SendGroupMessage(message models.Message) error
	GetPrivateConversation(userID1, userID2 string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)
	GetMessageByID(messageID string) (models.Message, error)
	GetMyConversations(userID string) ([]models.Conversation, error)
	CommentMessage(messageID, userID, comment string) error
	ForwardMessage(userID, messageID, toUserID, toGroupID string) error
	UncommentMessage(messageID string) error
	GetGroupByName(name string) (models.Group, error)
	IsMessageOwner(userID, messageID string) (bool, error)
	SetGroupPhoto(groupID, photoUrl string) error
	DeleteMessage(userID, messageID string) error

	// Conversations
	SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error)
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
}
