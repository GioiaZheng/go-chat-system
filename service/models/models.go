package models

// User represents the minimal public profile returned by the API.
// Internal-only fields such as usernames or photos are intentionally omitted.
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatarUri,omitempty"`
}

// Group captures the OpenAPI representation of a group, including membership
// and conversation linkage when available.
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatarUri,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversationId,omitempty"`
}

// GroupMember describes a user's role within a group that is returned to clients.
type GroupMember struct {
	UserID    string `json:"userId"`
	Name      string `json:"name,omitempty"`
	Role      string `json:"role,omitempty"`      // "admin" | "member"
	AvatarUrl string `json:"avatarUri,omitempty"` // user avatar
}

// Message is the internal representation used for both conversation messages
// and comment rows while hiding routing-only fields from the public surface.
type Message struct {
	ID             string `db:"id"`
	Content        string `db:"content"`
	FileURL        string `db:"file_url"`
	SenderID       string `db:"sender_id"`
	ConversationID string `db:"conversation_id"`
	CreatedAt      string `db:"created_at"`

	// Type contains the message classification (text, image, or file).
	Type   string `db:"type"`
	Status string `db:"status"`
	Read   bool   `db:"read"`

	// ReplyToID links to another message when the current one is a reply.
	ReplyToID *string `db:"reply_to_id"`

	// ReceiverID and GroupID are internal routing hints and are not exposed.
	ReceiverID string `db:"receiver_id"`
	GroupID    string `db:"group_id"`
}

// Conversation follows the OpenAPI contract and bundles participants and
// optional metadata about recency and creation times.
type Conversation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`           // "private" | "group"
	Name         string `json:"name,omitempty"` // group or derived name
	AvatarUrl    string `json:"avatarUri,omitempty"`
	Participants []User `json:"participants,omitempty"`

	LastMessage *Message `json:"lastMessage,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
}
