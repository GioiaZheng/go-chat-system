package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
)

func TestForwardMessageAllowsAuthorizedSourceAndTarget(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)
	targetID := createForwardTargetConversation(t, rt)
	req := httptest.NewRequest(http.MethodPost, "/messages/message-1/forwards", strings.NewReader(`{
		"conversationId":"`+targetID+`"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.forwardMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	messages, err := rt.db.GetMessagesByConversation(targetID, "", "", 10)
	if err != nil {
		t.Fatalf("GetMessagesByConversation returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("got %d forwarded messages, want 1", len(messages))
	}
	if messages[0].SenderID != "member-1" {
		t.Fatalf("SenderID = %q, want %q", messages[0].SenderID, "member-1")
	}
	if messages[0].Content != "seeded message" {
		t.Fatalf("Content = %q, want %q", messages[0].Content, "seeded message")
	}
}

func TestForwardMessageDeniesUnauthorizedSource(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)
	targetID := createForwardTargetConversation(t, rt)
	req := httptest.NewRequest(http.MethodPost, "/messages/message-1/forwards", strings.NewReader(`{
		"conversationId":"`+targetID+`"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.forwardMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertConversationMessageCount(t, rt, targetID, 0)
}

func TestForwardMessageDeniesUnauthorizedTarget(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)
	targetID := createForwardTargetConversation(t, rt)
	req := httptest.NewRequest(http.MethodPost, "/messages/message-1/forwards", strings.NewReader(`{
		"conversationId":"`+targetID+`"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.forwardMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-2"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertConversationMessageCount(t, rt, targetID, 0)
}

func createForwardTargetConversation(t *testing.T, rt *_router) string {
	t.Helper()
	conv, err := rt.db.StartConversation(context.Background(), "member-1", []string{"non-member"}, "forward target")
	if err != nil {
		t.Fatalf("StartConversation returned error: %v", err)
	}
	return conv.ID
}

func assertConversationMessageCount(t *testing.T, rt *_router, conversationID string, want int) {
	t.Helper()
	messages, err := rt.db.GetMessagesByConversation(conversationID, "", "", 10)
	if err != nil {
		t.Fatalf("GetMessagesByConversation returned error: %v", err)
	}
	if len(messages) != want {
		t.Fatalf("got %d messages, want %d", len(messages), want)
	}
}
