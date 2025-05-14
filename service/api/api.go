package api

import (
	"errors"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Config contains the dependencies required to create a new Router instance.
type Config struct {
	Logger   *logrus.Logger
	Database database.AppDatabase
}

// _router represents the internal router structure.
type _router struct {
	router     *httprouter.Router
	baseLogger *logrus.Logger
	db         database.AppDatabase
}

// New returns a new Router instance with the given configuration.
func New(cfg Config) (*_router, error) {
	// Validate the provided configuration
	if cfg.Logger == nil {
		return nil, errors.New("api: logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("api: database is required")
	}

	// Test database connection
	if err := cfg.Database.Ping(); err != nil {
		cfg.Logger.WithError(err).Error("database connection failed")
		return nil, errors.New("api: could not connect to database")
	}

	// Initialize the HTTP router
	router := httprouter.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true

	// Create router instance
	r := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// Register API routes
	r.RegisterRoutes()

	cfg.Logger.Info("API router initialized successfully")

	return r, nil
}

// Handler returns the HTTP handler for the router.
func (r *_router) Handler() http.Handler {
	return r.router
}

// validateToken 验证用户的 JWT 并返回用户 ID
func (rt *_router) validateToken(token string) (string, error) {
	// 假设从 token 中提取 userID，这里只是占位实现
	if token == "" {
		return "", errors.New("invalid token")
	}
	userID, err := rt.db.GetUserIDFromIdentifier(token)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to validate token")
		return "", err
	}
	return userID, nil
}
