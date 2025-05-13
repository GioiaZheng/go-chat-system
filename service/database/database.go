package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// AppDatabase interface
type AppDatabase interface {
	Ping() error
	Close() error

	// Auth
	CreateUser(user models.User) (models.User, error)
	AuthenticateUser(email string, password string) (models.User, string, error)

	// Users
	UpdateUserName(userID string, name string) error
	UpdateUserPhoto(userID string, photoPath string) error
	GetUserByID(userID string) (models.User, error)

	// Friends
	AddFriend(userID string, friendID string) error
	GetFriendsList(userID string) ([]models.Friend, error)
	AreFriends(userID1 string, userID2 string) bool

	// Groups
	CreateGroup(group models.Group) error
	AddMemberToGroup(groupID string, userID string, role string) error
	AddGroupMembers(groupID string, userIDs []string) error
	UpdateGroupName(groupID string, name string) error
	UpdateGroupPhoto(groupID string, photoPath string) error
	GetGroup(groupID string) (models.Group, error)
	GetGroupsList(userID string) ([]models.Group, error)
	LeaveGroup(groupID string, userID string) error

	// Messages
	SendMessageToUser(message models.Message) error
	SendMessageToGroup(message models.Message) error
	GetPrivateConversation(userID1 string, userID2 string) ([]models.Message, error)
	GetGroupConversation(groupID string) ([]models.Message, error)
	GetMessageByID(messageID string) (models.Message, error)
	GetMyConversations(userID string) ([]models.Conversation, error)
	CommentMessage(messageID string, userID string, comment string) error
	ForwardMessage(userID string, messageID string, toUserID string, toGroupID string) error
	UncommentMessage(messageID string) error
	GetGroupByName(name string) (models.Group, error)
	IsMessageOwner(userID string, messageID string) (bool, error)
	SendPrivateMessage(message models.Message) error
	SendGroupMessage(message models.Message) error
	SetGroupPhoto(groupID string, photoUrl string) error
	DeleteMessage(userID string, messageID string) error

	// Conversations
	SearchUsers(ctx context.Context, query string) ([]models.User, error)
	StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error)
}
