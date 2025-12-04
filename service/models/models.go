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
type Message struct {
	ID             string `json:"id"`
	Content        string `json:"content"`
	SenderID       string `json:"senderId"`
	ConversationID string `json:"conversationId,omitempty"`
	CreatedAt      string `json:"createdAt"`

	// Type can be:
	// - "text"
	// - "image"
	// - "file"
	// - "emoji" (for comment reactions)
	Type string `json:"type,omitempty"`

	// Status: "sent", "delivered", "read"
	Status string `json:"status,omitempty"`

	// Internal-only (NOT exposed in JSON)
	ReceiverID string `json:"-"`
	GroupID    string `json:"-"`

	// Optional for UI:
	// (not stored in DB — filled by API/FE if needed)
	FileUrl  string `json:"fileUrl,omitempty"`
	FileName string `json:"filename,omitempty"`
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
