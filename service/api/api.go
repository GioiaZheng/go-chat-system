package api

import (
	"errors"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Config collects the external dependencies required to build the API router.
type Config struct {
	Logger   *logrus.Logger
	Database database.AppDatabase
}

// _router is the internal HTTP router wrapper.
// It exposes a http.Handler via Handler() and provides route registration.
type _router struct {
	router     *httprouter.Router
	baseLogger *logrus.Logger
	db         database.AppDatabase
}

// New creates a new API router instance, validates dependencies,
// pings the database, configures the underlying httprouter,
// and registers all routes.
func New(cfg Config) (*_router, error) {
	if cfg.Logger == nil {
		return nil, errors.New("api: logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("api: database is required")
	}
	if err := cfg.Database.Ping(); err != nil {
		cfg.Logger.WithError(err).Error("database connection failed")
		return nil, errors.New("api: could not connect to database")
	}

	// Base router configuration.
	router := httprouter.New()
	// Redirects "/path" <-> "/path/" to reduce 404 surprises.
	router.RedirectTrailingSlash = true
	// Fixes small path issues like duplicate slashes or case differences.
	router.RedirectFixedPath = true

	// We handle OPTIONS ourselves via middleware.
	router.HandleOPTIONS = false

	r := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// Register all HTTP routes (public + protected).
	r.RegisterRoutes()

	cfg.Logger.Info("API router initialized successfully")
	return r, nil
}

// Handler returns the http.Handler to be mounted by main().
// We wrap the underlying router with CORS middleware so both
// actual requests and preflight requests are consistently handled.
func (r *_router) Handler() http.Handler {
	return withCORS(r.router)
}

// withCORS is a simple CORS middleware that:
// - sets the Access-Control-* headers on every response
// - short-circuits OPTIONS (preflight) with a 200 OK
// This keeps CORS centralized and avoids per-route boilerplate.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: For dev/demo we allow all origins. Tighten this for production.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "600")
		// Optional but good practice when proxies/CDNs are in the path:
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		// Handle preflight quickly and return.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Close is a no-op shutdown hook so main() can call apirouter.Close() safely.
// If you later add background workers, timers, open files, or any other
// resources that require cleanup, add the teardown code here.
func (rt *_router) Close() error {
	return nil
}
