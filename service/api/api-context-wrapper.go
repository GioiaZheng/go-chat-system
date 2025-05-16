package api

import (
	"context"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
	loggerKey    contextKey = "logger"
)

// GetUserIDFromContext extracts the user ID from the context.
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// SetUserIDInContext sets the user ID in the request context.
func SetUserIDInContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

// wrap is a middleware that adds context information to each request.
func (rt *_router) wrap(next func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 1. JWT Authentication
		token := r.Header.Get("Authorization")
		if token == "" {
			rt.baseLogger.Error("missing authorization token")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 2. Validate token and extract userID
		userID, err := rt.validateToken(token)
		if err != nil {
			rt.baseLogger.WithError(err).Error("invalid token")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 3. Generate Request UUID
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			http.Error(w, `{"code": 500, "message": "Internal Server Error"}`, http.StatusInternalServerError)
			return
		}

		// 4. Build Context with UserID, RequestID, and Logger
		ctx := context.WithValue(r.Context(), requestIDKey, reqUUID.String())
		ctx = context.WithValue(ctx, userIDKey, userID)

		logger := rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     reqUUID.String(),
			"remote-ip": r.RemoteAddr,
			"user_id":   userID,
		})
		ctx = context.WithValue(ctx, loggerKey, logger)

		// 5. Call the original handler with the enriched context
		next(w, r.WithContext(ctx), ps, reqcontext.RequestContext{
			UserID:  userID,
			ReqUUID: reqUUID,
			Logger:  logger,
		})
	}
}
