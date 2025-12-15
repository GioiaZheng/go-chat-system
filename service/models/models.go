package models

//
// ────────────────────────────────────────────────────────────────────────────────
//  USER
// ────────────────────────────────────────────────────────────────────────────────
//

// User is the public-facing user object.
// Internal DB fields (username, photo) are hidden.
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatarUri,omitempty"`
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  GROUP
// ────────────────────────────────────────────────────────────────────────────────
//

// Group returned to the frontend per OpenAPI: id, name, avatarUri, members, conversationId.
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatarUri,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversationId,omitempty"`
}

// GroupMember: public-facing structure.
type GroupMember struct {
	UserID    string `json:"userId"`
	Name      string `json:"name,omitempty"`
	Role      string `json:"role,omitempty"`      // "admin" | "member"
	AvatarUrl string `json:"avatarUri,omitempty"` // user avatar
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  MESSAGE (supports both normal messages & comments)
// ────────────────────────────────────────────────────────────────────────────────
//

// Message is used for:
// - conversation messages (OpenAPI Message)
// - comment rows (comment API returns Message-like shape)
// Only ReceiverID / GroupID are internal-only.
// Message is the internal database model for messages.
type Message struct {
	ID             string  `db:"id"`
	Content        string  `db:"content"`
	SenderID       string  `db:"sender_id"`
	ConversationID string  `db:"conversation_id"`
	CreatedAt      string  `db:"created_at"`

	// message type: text | image | file
	Type   string `db:"type"`
	Status string `db:"status"`

        // Messaging-style reply that points to another message
	ReplyToID *string `db:"reply_to_id"`

	// Internal-only routing fields
	ReceiverID string `db:"receiver_id"`
	GroupID    string `db:"group_id"`
}


//
// ────────────────────────────────────────────────────────────────────────────────
//  CONVERSATION
// ────────────────────────────────────────────────────────────────────────────────
//

// Conversation matches OpenAPI Conversation:
// required: id, type, participants
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
