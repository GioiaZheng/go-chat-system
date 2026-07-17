package database

import (
	"path/filepath"
	"testing"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

func TestGetAllMessagesReturnsInsertedMessage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "messages.db")
	sqlDB, err := OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	appDB, err := New(sqlDB)
	if err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	defer appDB.Close()

	_, err = appDB.c.Exec(`INSERT INTO users (id, name) VALUES (?, ?), (?, ?)`,
		"sender-1", "sender",
		"receiver-1", "receiver",
	)
	if err != nil {
		t.Fatalf("insert users: %v", err)
	}

	message := models.Message{
		ID:         "message-1",
		Content:    "hello from sqlite",
		Type:       "text",
		Status:     "sent",
		Read:       true,
		SenderID:   "sender-1",
		ReceiverID: "receiver-1",
		CreatedAt:  "2026-05-13T12:00:00Z",
	}
	if err := appDB.SendPrivateMessage(message); err != nil {
		t.Fatalf("insert message: %v", err)
	}

	messages, err := appDB.GetAllMessages()
	if err != nil {
		t.Fatalf("get all messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(messages))
	}

	got := messages[0]
	if got.ID != message.ID {
		t.Errorf("ID = %q, want %q", got.ID, message.ID)
	}
	if got.Content != message.Content {
		t.Errorf("Content = %q, want %q", got.Content, message.Content)
	}
	if got.Type != message.Type {
		t.Errorf("Type = %q, want %q", got.Type, message.Type)
	}
	if got.Status != message.Status {
		t.Errorf("Status = %q, want %q", got.Status, message.Status)
	}
	if got.Read != message.Read {
		t.Errorf("Read = %t, want %t", got.Read, message.Read)
	}
	if got.SenderID != message.SenderID {
		t.Errorf("SenderID = %q, want %q", got.SenderID, message.SenderID)
	}
	if got.ReceiverID != message.ReceiverID {
		t.Errorf("ReceiverID = %q, want %q", got.ReceiverID, message.ReceiverID)
	}
	if got.CreatedAt != message.CreatedAt {
		t.Errorf("CreatedAt = %q, want %q", got.CreatedAt, message.CreatedAt)
	}
}
