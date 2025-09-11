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

	router := httprouter.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true

	router.HandleOPTIONS = false
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "1")
		w.WriteHeader(http.StatusOK)
	})

	r := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}
	r.RegisterRoutes()
	cfg.Logger.Info("API router initialized successfully")
	return r, nil
}

func (r *_router) Handler() http.Handler {
	return withCORS(r.router)
}

// ---- CORS middleware ----
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "1")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
