package models

// --- User ---
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Password  string `json:"-"`                    // not exposed
	AvatarUrl string `json:"avatar_url,omitempty"` // optional, exposed
	Photo     string `json:"-"`                    // internal file path, not exposed
}

// --- Group ---
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatar_url,omitempty"`
	CreatedAt      string        `json:"created_at,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversation_id,omitempty"`
}

// --- GroupMember ---
type GroupMember struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"username"`
	Role      string `json:"role,omitempty"`       // "admin" | "member"
	AvatarUrl string `json:"avatar_url,omitempty"` // convenience
}

// --- Message ---
type Message struct {
	ID             string `json:"id"`
	Content        string `json:"content"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id,omitempty"`
	GroupID        string `json:"group_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	CreatedAt      string `json:"created_at"`

	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// --- Conversation ---
type Conversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	LastMessage string `json:"last_message,omitempty"`
}
