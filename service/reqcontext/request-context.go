// Package reqcontext defines the per-request context populated by the
// middleware in api-context-wrapper.go. Each field should be treated as
// request-scoped unless explicitly documented otherwise.
// Related files: service/api/api-context-wrapper.go, service/api/api.go.
package reqcontext

import (
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

// RequestContext aggregates request-dependent parameters attached by middleware.
type RequestContext struct {
	// ReqUUID is the unique identifier associated with the current request.
	ReqUUID uuid.UUID

	// Logger is a structured logger scoped to the current request.
	Logger logrus.FieldLogger

	// UserID is the identifier of the authenticated user.
	UserID string

	// Identifier is the authentication token provided by the client.
	Identifier string
}
