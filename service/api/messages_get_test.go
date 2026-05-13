package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
)

func TestGetMessagesAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)

	req := httptest.NewRequest(http.MethodGet, "/messages?conversationId=conversation-1", nil)
	rr := httptest.NewRecorder()

	rt.getMessages(rr, req, nil, reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "message-1") {
		t.Fatalf("response body does not contain seeded message: %s", rr.Body.String())
	}
}

func TestGetMessagesDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)

	req := httptest.NewRequest(http.MethodGet, "/messages?conversationId=conversation-1", nil)
	rr := httptest.NewRecorder()

	rt.getMessages(rr, req, nil, reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestGetMessagesNonMemberDoesNotMarkConversationRead(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)

	req := httptest.NewRequest(http.MethodGet, "/messages?conversationId=conversation-1", nil)
	rr := httptest.NewRecorder()

	rt.getMessages(rr, req, nil, reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}

	msg, err := rt.db.GetMessageByID("message-1")
	if err != nil {
		t.Fatalf("GetMessageByID returned error: %v", err)
	}
	if msg.Read {
		t.Fatal("message was marked read for a non-member request")
	}
}

func seedConversationMessage(t *testing.T, rt *_router, id string, read bool) {
	t.Helper()

	msg := models.Message{
		ID:             id,
		Content:        "seeded message",
		SenderID:       "member-2",
		ConversationID: "conversation-1",
		CreatedAt:      "2026-05-13T12:00:00Z",
		Type:           "text",
		Status:         "sent",
		Read:           read,
	}
	if err := rt.db.SendMessageToConversation(msg); err != nil {
		t.Fatalf("SendMessageToConversation returned error: %v", err)
	}
}
