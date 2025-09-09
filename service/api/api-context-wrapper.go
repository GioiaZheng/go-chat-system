package api

import (
	"errors"
	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	// "context"
)

// Custom type for context keys to avoid collisions
type contextKey string

const UserIDKey = contextKey("userId")

var (
	ErrNoAuthHeader      = errors.New("authorization header required")
	ErrInvalidAuthFormat = errors.New("invalid authorization header format")
	ErrNoToken           = errors.New("token required")
)

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		identifier, err := rt.extractToken(r)
		if err != nil {
			rt.baseLogger.WithError(err).Warn("authentication failed")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Get user ID from identifier
		userID, err := rt.db.GetUserIDFromIdentifier(identifier)
		if err != nil {
			rt.baseLogger.WithError(err).Warn("invalid identifier")
			http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}

		var ctx = reqcontext.RequestContext{
			ReqUUID:    reqUUID,
			UserID:     userID,
			Identifier: identifier,
		}

		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
			"user":      userID,
		})

		fn(w, r, ps, ctx)
	}
}

// // authenticate checks the Authorization header and returns the user ID
// extractToken gets the token from the Authorization header
func (rt *_router) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidAuthFormat
	}

	token := parts[1]
	if token == "" {
		return "", ErrNoToken
	}

	return token, nil
}
