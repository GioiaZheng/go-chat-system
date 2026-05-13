package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func TestGetMessageByIDAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)

	req := httptest.NewRequest(http.MethodGet, "/messages/message-1", nil)
	rr := httptest.NewRecorder()

	rt.getMessageByID(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "message-1") {
		t.Fatalf("response body does not contain seeded message: %s", rr.Body.String())
	}
}

func TestGetMessageByIDDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)

	req := httptest.NewRequest(http.MethodGet, "/messages/message-1", nil)
	rr := httptest.NewRecorder()

	rt.getMessageByID(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestGetMessageCommentsAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessageWithComment(t, rt)

	req := httptest.NewRequest(http.MethodGet, "/messages/message-1/comment", nil)
	rr := httptest.NewRecorder()

	rt.getMessageComments(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "seeded comment") {
		t.Fatalf("response body does not contain seeded comment: %s", rr.Body.String())
	}
}

func TestGetMessageCommentsDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessageWithComment(t, rt)

	req := httptest.NewRequest(http.MethodGet, "/messages/message-1/comment", nil)
	rr := httptest.NewRecorder()

	rt.getMessageComments(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestCommentMessageAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)
	req := httptest.NewRequest(http.MethodPost, "/messages/message-1/comment", strings.NewReader(`{
		"type":"text",
		"content":"new comment"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.commentMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	assertCommentCount(t, rt, "message-1", 1)
}

func TestCommentMessageDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessage(t, rt, "message-1", false)
	req := httptest.NewRequest(http.MethodPost, "/messages/message-1/comment", strings.NewReader(`{
		"type":"text",
		"content":"new comment"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.commentMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertCommentCount(t, rt, "message-1", 0)
}

func TestUncommentMessageAllowsConversationMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessageWithComment(t, rt)
	req := httptest.NewRequest(http.MethodDelete, "/messages/message-1/comment", nil)
	rr := httptest.NewRecorder()

	rt.uncommentMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "member-1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertCommentCount(t, rt, "message-1", 0)
}

func TestUncommentMessageDeniesConversationNonMember(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	seedConversationMessageWithComment(t, rt)
	req := httptest.NewRequest(http.MethodDelete, "/messages/message-1/comment", nil)
	rr := httptest.NewRecorder()

	rt.uncommentMessage(rr, req, messageIDParams("message-1"), reqcontext.RequestContext{UserID: "non-member"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertCommentCount(t, rt, "message-1", 1)
}

func seedConversationMessageWithComment(t *testing.T, rt *_router) {
	t.Helper()
	seedConversationMessage(t, rt, "message-1", false)
	if err := rt.db.CommentMessage("message-1", "member-1", "text", "seeded comment"); err != nil {
		t.Fatalf("CommentMessage returned error: %v", err)
	}
}

func assertCommentCount(t *testing.T, rt *_router, messageID string, want int) {
	t.Helper()
	comments, err := rt.db.GetMessageComments(messageID)
	if err != nil {
		t.Fatalf("GetMessageComments returned error: %v", err)
	}
	if len(comments) != want {
		t.Fatalf("got %d comments, want %d", len(comments), want)
	}
}

func messageIDParams(id string) httprouter.Params {
	return httprouter.Params{{Key: "id", Value: id}}
}
