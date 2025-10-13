package models

// --- User ---
// Field names keep backward-compatibility for DB code (AvatarUrl).
// JSON tags follow OpenAPI (avatarUri).
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Password  string `json:"-"`                     // internal only, never exposed
	AvatarUrl string `json:"avatarUri,omitempty"`   // exposed as avatarUri
	Photo     string `json:"-"`                     // internal file path, never exposed
}

// --- Group ---
// Keep AvatarUrl for DB; expose as avatarUri in JSON.
// conversationId aligns with OpenAPI.
type Group struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	AvatarUrl      string        `json:"avatarUri,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	ConversationID string        `json:"conversationId,omitempty"`
}

// --- GroupMember ---
// Convenience avatar also exposed as avatarUri.
type GroupMember struct {
	UserID    string `json:"userId"`
	UserName  string `json:"username"`
	Role      string `json:"role,omitempty"`       // "admin" | "member"
	AvatarUrl string `json:"avatarUri,omitempty"`  // convenience
}

// --- Message ---
// OpenAPI is conversation-centric; we still keep ReceiverID and GroupID
// for DB compatibility but hide them from JSON with json:"-".
// createdAt / senderId / conversationId / type / status follow spec.
type Message struct {
	ID             string `json:"id"`
	Content        string `json:"content"`
	SenderID       string `json:"senderId"`
	ConversationID string `json:"conversationId,omitempty"`
	CreatedAt      string `json:"createdAt"`

	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`

	// DB-only legacy fields (not exposed)
	ReceiverID string `json:"-"`
	GroupID    string `json:"-"`
}

// --- Conversation ---
// OpenAPI allows lastMessage to be a Message object; make it pointer for optionality.
type Conversation struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AvatarUrl   string   `json:"avatarUri,omitempty"`
	LastMessage *Message `json:"lastMessage,omitempty"`
}
