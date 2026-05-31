package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
)

func TestDoLoginCreatesUserAndReturnsToken(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"Diana"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.doLogin(rr, req, nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var got authEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Data.User.ID == "" {
		t.Fatal("expected created user id")
	}
	if got.Data.Token != got.Data.User.ID {
		t.Fatalf("token = %q, want user id %q", got.Data.Token, got.Data.User.ID)
	}
	if got.Data.User.Name != "Diana" {
		t.Fatalf("user name = %q, want Diana", got.Data.User.Name)
	}
}

func TestDoLoginReturnsExistingUserToken(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"member one"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.doLogin(rr, req, nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var got authEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Data.User.ID != "member-1" {
		t.Fatalf("user id = %q, want member-1", got.Data.User.ID)
	}
	if got.Data.Token != "member-1" {
		t.Fatalf("token = %q, want member-1", got.Data.Token)
	}
}

func TestDoLoginRejectsInvalidName(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rt.doLogin(rr, req, nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestSearchUsersRequiresQueryAndExcludesCaller(t *testing.T) {
	rt := newConversationAuthorizationTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/users/search?q=member", nil)
	rr := httptest.NewRecorder()

	rt.searchUsers(rr, req, nil, requestContextForTest("member-1"))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	if strings.Contains(body, "member-1") {
		t.Fatalf("search response includes caller: %s", body)
	}
	if !strings.Contains(body, "member-2") {
		t.Fatalf("search response does not include matching peer: %s", body)
	}
}

func requestContextForTest(userID string) reqcontext.RequestContext {
	return reqcontext.RequestContext{UserID: userID, Identifier: userID}
}
