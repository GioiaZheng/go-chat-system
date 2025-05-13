package api

import (
	"context"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// --- 内部类型：用于从context读取userID ---
type contextKey string

const userIDKey contextKey = "userID"

// GetUserIDFromContext extracts the userID from context
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// SetUserIDInContext saves the userID into context
func SetUserIDInContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

// --- wrap函数 ---
// 用来包裹所有需要认证、记录日志的 handler
// 负责生成reqID，注入到 context 里，并交给下游处理
func (rt *_router) wrap(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 生成Request UUID
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 构建自定义RequestContext（可以存放日志对象）
		ctx := context.WithValue(r.Context(), reqcontext.RequestIDKey, reqUUID.String())

		logger := rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     reqUUID.String(),
			"remote-ip": r.RemoteAddr,
		})
		ctx = context.WithValue(ctx, reqcontext.LoggerKey, logger)

		// 将新的Context挂载到Request上
		r = r.WithContext(ctx)

		// 调用原始 handler
		next(w, r, ps)
	}
}
