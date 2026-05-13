package api

import (
	"database/sql"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/GioiaZheng/Wasa_proj/service/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func TestRequireConversationMemberAllowsMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)

	if err := rt.requireConversationMember("member-1", "conversation-1"); err != nil {
		t.Fatalf("requireConversationMember returned error for member: %v", err)
	}
}

func TestRequireConversationMemberDeniesNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)

	err := rt.requireConversationMember("non-member", "conversation-1")
	if !errors.Is(err, ErrConversationNotMember) {
		t.Fatalf("requireConversationMember error = %v, want %v", err, ErrConversationNotMember)
	}
}

func TestRequireConversationMemberDeniesEmptyUserID(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)

	err := rt.requireConversationMember(" ", "conversation-1")
	if !errors.Is(err, ErrConversationUserRequired) {
		t.Fatalf("requireConversationMember error = %v, want %v", err, ErrConversationUserRequired)
	}
}

func TestRequireConversationMemberDeniesEmptyConversationID(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)

	err := rt.requireConversationMember("member-1", " ")
	if !errors.Is(err, ErrConversationIDRequired) {
		t.Fatalf("requireConversationMember error = %v, want %v", err, ErrConversationIDRequired)
	}
}

func newConversationAuthorizationTestRouter(t *testing.T) *_router {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "conversation-auth.db"))
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	appDB, err := database.New(sqlDB)
	if err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	t.Cleanup(func() { _ = appDB.Close() })

	_, err = sqlDB.Exec(`
		INSERT INTO users (id, name) VALUES
			('member-1', 'member one'),
			('member-2', 'member two'),
			('non-member', 'non member');
		INSERT INTO conversations (id, name) VALUES ('conversation-1', 'test conversation');
		INSERT INTO conversation_members (conversation_id, user_id) VALUES
			('conversation-1', 'member-1'),
			('conversation-1', 'member-2');
	`)
	if err != nil {
		t.Fatalf("seed conversation membership: %v", err)
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &_router{baseLogger: logger, db: appDB}
}
