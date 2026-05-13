// api.go constructs the HTTP router for the service API and centralizes shared
// dependencies such as logging and database connectivity.
// Related files: service/api/api-handler.go, service/api/api-context-wrapper.go, service/database/app_database.go.
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
// CORS is configured at the HTTP server entrypoint in cmd/webapi/cors.go.
func (r *_router) Handler() http.Handler {
	return r.router
}

// Close is a no-op shutdown hook so main() can call apirouter.Close() safely.
// If you later add background workers, timers, open files, or any other
// resources that require cleanup, add the teardown code here.
func (rt *_router) Close() error {
	return nil
}
