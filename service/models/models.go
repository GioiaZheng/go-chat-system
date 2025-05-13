package models

// User structure
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	AvatarUrl string `json:"avatarUrl"`
	Photo     string `json:"photo,omitempty"`
	Gender    string `json:"gender"`
}

// Friend structure
type Friend struct {
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	AvatarUrl string `json:"avatarUrl"`
}

// Group structure
type Group struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	AvatarUrl string        `json:"avatarUrl,omitempty"`
	CreatedAt string        `json:"createdAt"`
	Members   []GroupMember `json:"members"`
}

// GroupMember structure
type GroupMember struct {
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	Role      string `json:"role"`
	AvatarUrl string `json:"avatarUrl"`
}

// Message structure
type Message struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId,omitempty"`
	GroupID    string `json:"groupId,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// Conversation structure
type Conversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatarUrl"`
	LastMessage string `json:"lastMessage"`
}
