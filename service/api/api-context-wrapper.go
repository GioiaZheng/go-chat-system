// api-context-wrapper.go enriches httprouter handlers with request metadata and
// authentication details so business handlers receive consistent context.
package api

import (
	"errors"
	"net/http"
	"strings"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var (
	// Returned when the Authorization header is missing entirely.
	ErrNoAuthHeader = errors.New("missing Authorization header")
	// Returned when the token is present but empty after trimming.
	ErrNoToken = errors.New("empty token")
)

// httpRouterHandler is the handler signature expected by rt.wrap.
// It receives an extra RequestContext already populated with request metadata and auth info.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap handles request-scoped concerns before calling the actual business handler:
// 1) Generate a per-request UUID for tracing.
// 2) Extract the auth token from the "Authorization" header.
// 3) Build a RequestContext (logger + user identity) for downstream use.
//
// Authentication model used in this assignment:
//   - Login/Register issue an "opaque token" equal to the user ID.
//   - For convenience and to match the assignment scripts / existing clients,
//     both "Authorization: Bearer <token>" and a bare "<token>" are accepted.
//   - We treat "<token>" directly as the user ID to avoid a second lookup.
//     (If you prefer, enable the optional DB existence check below.)
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("failed to generate request UUID")
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		token, err := extractTokenFromHeader(r)
		if err != nil {
			// We log at Warn level because this is a client-side error in most cases.
			rt.baseLogger.WithError(err).Warn("failed to parse Authorization header")
			rt.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing authentication token")
			return
		}

		// In this simplified model: token == userID.
		userID := token

		// OPTIONAL: Verify the user still exists in DB.
		// if _, err := rt.db.GetUserByID(userID); err != nil {
		// 	rt.baseLogger.WithError(err).Warn("user not found for provided token")
		// 	rt.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired authentication token")
		// 	return
		// }

		ctx := reqcontext.RequestContext{
			ReqUUID:    reqUUID,
			UserID:     userID,
			Identifier: userID, // Keep identical for compatibility with existing code paths.
		}
		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
			"user":      userID,
		})

		fn(w, r, ps, ctx)
	}
}

// extractTokenFromHeader extracts the auth token from the "Authorization" header.
// Accepted formats:
//  1. "Authorization: Bearer <token>"
//  2. "Authorization: <token>" (bare token, kept for assignment/client compatibility)
func extractTokenFromHeader(r *http.Request) (string, error) {
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if raw == "" {
		return "", ErrNoAuthHeader
	}

	low := strings.ToLower(raw)
	if strings.HasPrefix(low, "bearer ") {
		token := strings.TrimSpace(raw[len("Bearer "):])
		if token == "" {
			return "", ErrNoToken
		}
		return token, nil
	}

	// Fallback: treat the whole header as a bare token (must not be empty after trimming).
	if raw == "" {
		return "", ErrNoToken
	}
	return raw, nil
}
