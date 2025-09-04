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

	// （可留可去）全局 OPTIONS 兜底
	router.HandleOPTIONS = false
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "1")
		w.WriteHeader(http.StatusOK)
	})

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
// 关键：用 CORS 中间件包住整个路由
func (r *_router) Handler() http.Handler {
	return withCORS(r.router)
}

// validateToken 验证用户的 token（此处 token = user.ID）
func (rt *_router) validateToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("invalid token")
	}
	user, err := rt.db.GetUserByID(token)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to validate token")
		return "", errors.New("invalid token")
	}
	return user.ID, nil
}

// ---- CORS middleware ----
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 通用 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "1")

		// 预检请求直接返回
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 其余交给下游
		next.ServeHTTP(w, r)
	})
}
