package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// AppDatabase is the public interface consumed by the API layer.
// appdbimpl provides the concrete implementation and must satisfy this interface.
type AppDatabase interface {
	// lifecycle
	Close() error
	Ping() error

	// users
	GetUserByID(userID string) (models.User, error)
	GetUser(userID string) (models.User, error) // some code paths call GetUser
	GetUserIDFromIdentifier(identifier string) (string, error)
	CheckUserExists(username string) (bool, error)
	GetUserByCredentials(username, password string) (models.User, error)
	CreateUser(user models.User, password string) (models.User, error)
	UpdateUserName(userID, username string) error
	UpdateUserPhoto(userID, avatarURL string) error
	SearchUsers(ctx context.Context, excludeUserID, query string) ([]models.User, error)

	// groups
	CreateGroup(group models.Group) error
	AddGroupMembers(groupID string, memberIDs []string) error
	GetGroup(groupID string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID, url string) error
	LeaveGroup(groupID, userID string) error
	GetGroupMembers(groupID string) ([]models.User, error)

	// conversations
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
	GetMyConversations(userID string) ([]models.Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)

	// messages
	SendMessageToConversation(m models.Message) error
	SendPrivateMessage(m models.Message) error
	GetAllMessages() ([]models.Message, error)
	GetMessageByID(id string) (models.Message, error)
	GetPrivateConversation(me, other string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)
	ForwardMessage(userID, msgID, toUserID, toGroupID string) error
	IsMessageOwner(userID, msgID string) (bool, error)
	DeleteMessage(userID, msgID string) error

	// comments
	GetMessageComments(messageID string) ([]models.Message, error)
	CommentMessage(messageID, authorID, content string) error
	UncommentMessage(messageID string) error
}
