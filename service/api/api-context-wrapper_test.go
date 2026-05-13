package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func TestAuthWrapperRejectsMissingAuthorizationHeader(t *testing.T) {
	rt := newAuthWrapperTestRouter()
	called := false

	handler := rt.wrap(func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	handler(rr, req, nil)

	if called {
		t.Fatal("wrapped handler was called without an Authorization header")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthWrapperRejectsEmptyBearerToken(t *testing.T) {
	rt := newAuthWrapperTestRouter()
	called := false

	handler := rt.wrap(func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()

	handler(rr, req, nil)

	if called {
		t.Fatal("wrapped handler was called with an empty Bearer token")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthWrapperAllowsValidBearerToken(t *testing.T) {
	rt := newAuthWrapperTestRouter()
	const token = "user-123"
	var got reqcontext.RequestContext
	called := false

	handler := rt.wrap(func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
		called = true
		got = ctx
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler(rr, req, nil)

	if !called {
		t.Fatal("wrapped handler was not called with a valid Bearer token")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got.UserID != token {
		t.Fatalf("UserID = %q, want %q", got.UserID, token)
	}
	if got.Identifier != token {
		t.Fatalf("Identifier = %q, want %q", got.Identifier, token)
	}
}

func TestAuthWrapperAllowsValidBareToken(t *testing.T) {
	rt := newAuthWrapperTestRouter()
	const token = "user-456"
	var got reqcontext.RequestContext
	called := false

	handler := rt.wrap(func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
		called = true
		got = ctx
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	handler(rr, req, nil)

	if !called {
		t.Fatal("wrapped handler was not called with a valid bare token")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got.UserID != token {
		t.Fatalf("UserID = %q, want %q", got.UserID, token)
	}
	if got.Identifier != token {
		t.Fatalf("Identifier = %q, want %q", got.Identifier, token)
	}
}

func newAuthWrapperTestRouter() *_router {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &_router{baseLogger: logger}
}
