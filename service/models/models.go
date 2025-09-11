package models

// --- User ---
// JSON fields aligned to OpenAPI (snake_case).
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email"`
	Password  string `json:"-"`                    // not exposed
	AvatarUrl string `json:"avatar_url,omitempty"` // matches api.yaml
	Photo     string `json:"-"`                    // internal storage path if any
	Gender    string `json:"gender,omitempty"`     // "male" | "female" | "unspecified"
}

// --- Group ---
// JSON fields aligned to OpenAPI (snake_case).
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatar_url,omitempty"`
	CreatedAt      string        `json:"created_at,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversation_id,omitempty"`
}

// --- GroupMember ---
// JSON fields aligned to OpenAPI (snake_case).
type GroupMember struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"username"`
	Role      string `json:"role,omitempty"`       // "admin" | "member"
	AvatarUrl string `json:"avatar_url,omitempty"` // optional convenience
}

// --- Message ---
// JSON fields aligned to OpenAPI (snake_case).
// Added optional Type/Status to better match the spec; they are omitempty so legacy code won't break.
type Message struct {
	ID             string `json:"id"`
	Content        string `json:"content"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id,omitempty"`
	GroupID        string `json:"group_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	CreatedAt      string `json:"created_at"`

	// Optional fields to align with OpenAPI (enum examples: type: text|image|video|file; status: sending|sent|delivered|read|failed)
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// --- Conversation ---
// Internal summary model; JSON fields aligned to OpenAPI (snake_case).
type Conversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	LastMessage string `json:"last_message,omitempty"`
}
