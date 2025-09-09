package reqcontext

import (
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

// RequestContext 是每个请求的上下文
type RequestContext struct {
	// ReqUUID 是请求唯一 ID
	ReqUUID uuid.UUID

	// Logger 是请求的定制日志记录器
	Logger logrus.FieldLogger

	// UserID 是发起请求的用户 ID
	UserID string

	// Identifier 是用户使用的认证令牌
	Identifier string
}
