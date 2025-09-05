package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Internal context keys used to store request-scoped values.
// Keep them unexported to avoid collisions outside this package.
type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
	loggerKey    contextKey = "logger"
)

// errorPayload is a unified JSON shape for error responses.
// Using a struct ensures stable field names and easy encoding.
type errorPayload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// -------------------------- JSON Helpers --------------------------

// writeJSON writes a JSON response with the provided status code.
// It sets the Content-Type header and encodes the given value.
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// sendError writes a structured error JSON while preserving backward compatibility
// with existing call sites across the project. Optional details can be attached
// (they will be omitted if empty).
func (rt *_router) sendError(w http.ResponseWriter, status int, msg string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var det interface{}
	switch len(details) {
	case 0:
		det = nil
	case 1:
		det = details[0]
	default:
		det = details
	}

	_ = json.NewEncoder(w).Encode(errorPayload{
		Code:    status,
		Message: msg,
		Details: det,
	})
}

// ----------------------- Request Context Utils --------------------

// GetUserIDFromContext returns the authenticated userID stored in the request context.
// It is safe to call with a nil context and will return an empty string if not present.
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// SetUserIDInContext attaches a userID into the request context and returns a new *http.Request.
// This is useful for tests or internal rewrites where you need to inject identity.
func SetUserIDInContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

// ----------------------- Auth Wrap Middleware ---------------------

// wrap is the auth+context middleware used for all protected routes.
// It performs the following steps:
//  1) Parse "Authorization: Bearer <token>" header.
//  2) Validate the token (in this assignment we treat the token as userID).
//  3) Create a per-request UUID.
//  4) Attach requestID, userID and a request-scoped logger into the context.
//  5) Call the next handler with an enriched reqcontext.RequestContext.
func (rt *_router) wrap(
	next func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext),
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 1) Parse "Authorization: Bearer <token>"
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			rt.baseLogger.Error("missing or malformed authorization header")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			rt.baseLogger.Error("empty bearer token")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 2) Validate token (current behavior: token == userID).
		//    If you later switch to real JWT, only this method needs to change.
		userID, err := rt.validateToken(token)
		if err != nil || userID == "" {
			if err != nil {
				rt.baseLogger.WithError(err).Error("invalid token")
			} else {
				rt.baseLogger.Error("invalid token: empty userID")
			}
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 3) Generate a per-request UUID for tracing and correlation.
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("failed to generate request UUID")
			http.Error(w, `{"code": 500, "message": "Internal Server Error"}`, http.StatusInternalServerError)
			return
		}

		// 4) Enrich the context with requestID, userID and a scoped logger.
		ctx := context.WithValue(r.Context(), requestIDKey, reqUUID.String())
		ctx = context.WithValue(ctx, userIDKey, userID)
		logger := rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     reqUUID.String(),
			"remote_ip": r.RemoteAddr,
			"user_id":   userID,
		})
		ctx = context.WithValue(ctx, loggerKey, logger)

		// 5) Call the next handler with our typed reqcontext, which mirrors what
		//    you already use across the project.
		next(w, r.WithContext(ctx), ps, reqcontext.RequestContext{
			UserID:  userID,
			ReqUUID: reqUUID,
			Logger:  logger,
		})
	}
}

// readJSON reads and validates a JSON request body into dst.
// Features:
//  - Limits body size (default: 1MB) to avoid abuse.
//  - Disallows unknown fields to catch client typos early.
//  - Returns clear errors for empty body or malformed JSON.
func readJSON(r *http.Request, dst interface{}) error {
	const maxBody = 1 << 20 // 1 MiB

	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, maxBody))
	dec.DisallowUnknownFields()

	// Decode the first JSON object
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Ensure there's no trailing garbage after the first JSON object
	// (e.g., "{}{}" or "{} extra")
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: multiple JSON values in body")
		}
		return fmt.Errorf("invalid JSON (trailing content): %w", err)
	}

	return nil
}
