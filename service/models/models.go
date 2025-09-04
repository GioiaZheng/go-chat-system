package models

// --- User ---
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AvatarUrl string `json:"avatarUrl,omitempty"`
	Photo     string `json:"-"`
	Gender    string `json:"gender,omitempty"`
}

// --- Group ---
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatarUrl,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversationId,omitempty"`
}

// --- GroupMember ---
type GroupMember struct {
	UserID    string `json:"userId"`
	UserName  string `json:"username"`
	Role      string `json:"role,omitempty"`
	AvatarUrl string `json:"avatarUrl,omitempty"`
}

// --- Message ---
type Message struct {
	ID             string `json:"id"`
	Content        string `json:"content"`
	SenderID       string `json:"senderId"`
	ReceiverID     string `json:"receiverId,omitempty"`
	GroupID        string `json:"groupId,omitempty"`
	ConversationID string `json:"conversationId,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

// --- Conversation ---
type Conversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatarUrl,omitempty"`
	LastMessage string `json:"lastMessage,omitempty"`
}
