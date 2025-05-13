package models

// User structure
// Matches the OpenAPI definition for User
// Required fields: id, username, email, gender

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarUrl"`
	Photo     string `json:"photo"`
	Gender    string `json:"gender"`
	Password  string `json:"-"`
}

// Friend structure
// Matches the OpenAPI definition for Friend
// Required fields: userId, userName

type Friend struct {
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	AvatarUrl string `json:"avatarUrl"`
}

// Group structure
// Matches the OpenAPI definition for Group
// Required fields: id, name, members

type Group struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Members []GroupMember `json:"members"`
}

// GroupMember structure
// Matches the OpenAPI definition for GroupMember
// Required fields: userId, role

type GroupMember struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// Message structure
// Matches the OpenAPI definition for Message
// Required fields: content, senderId, createdAt

type Message struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId,omitempty"`
	GroupID    string `json:"groupId,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// Conversation structure
// Matches the OpenAPI definition for Conversation
// Required fields: id, name

type Conversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatarUrl"`
	LastMessage string `json:"lastMessage"`
}
