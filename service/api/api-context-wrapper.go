package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

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

type errorPayload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ---------- 通用 JSON 工具 ----------

func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	_ = writeJSON(w, statusCode, map[string]string{"error": message})
}

func readJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// sendError：项目里已有大量调用，保持兼容
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

// ---------- 请求上下文辅助 ----------

func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

func SetUserIDInContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

// ---------- wrap 中间件：为所有受保护路由注入 ctx ----------

func (rt *_router) wrap(next func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 1) 解析 Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			rt.baseLogger.Error("missing or malformed authorization header")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			rt.baseLogger.Error("empty bearer token")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 2) 校验 token（当前实现：token 即 userID）
		userID, err := rt.validateToken(token)
		if err != nil {
			rt.baseLogger.WithError(err).Error("invalid token")
			http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// 3) 生成请求 ID
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			http.Error(w, `{"code": 500, "message": "Internal Server Error"}`, http.StatusInternalServerError)
			return
		}

		// 4) 注入上下文
		ctx := context.WithValue(r.Context(), requestIDKey, reqUUID.String())
		ctx = context.WithValue(ctx, userIDKey, userID)

		logger := rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     reqUUID.String(),
			"remote-ip": r.RemoteAddr,
			"user_id":   userID,
		})
		ctx = context.WithValue(ctx, loggerKey, logger)

		// 5) 调用下游处理，附带我们自己的 reqcontext
		next(w, r.WithContext(ctx), ps, reqcontext.RequestContext{
			UserID:  userID,
			ReqUUID: reqUUID,
			Logger:  logger,
		})
	}
}
