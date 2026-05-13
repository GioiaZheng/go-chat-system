package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
)

func TestSendMessageAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(`{
		"conversationId":"conversation-1",
		"content":"hello from a member"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.sendMessage(rr, req, nil, reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	messages, err := rt.db.GetMessagesByConversation("conversation-1", "", "", 10)
	if err != nil {
		t.Fatalf("GetMessagesByConversation returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(messages))
	}
	if messages[0].SenderID != "member-1" {
		t.Fatalf("SenderID = %q, want %q", messages[0].SenderID, "member-1")
	}
	if messages[0].Content != "hello from a member" {
		t.Fatalf("Content = %q, want %q", messages[0].Content, "hello from a member")
	}
}

func TestSendMessageDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(`{
		"conversationId":"conversation-1",
		"content":"hello from a non-member"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.sendMessage(rr, req, nil, reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}

	messages, err := rt.db.GetMessagesByConversation("conversation-1", "", "", 10)
	if err != nil {
		t.Fatalf("GetMessagesByConversation returned error: %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("got %d messages, want 0", len(messages))
	}
}
