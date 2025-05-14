package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GioiaZheng/Wasa_proj/service/api"
	"github.com/GioiaZheng/Wasa_proj/service/database"
	"github.com/GioiaZheng/Wasa_proj/service/globaltime"
	"github.com/ardanlabs/conf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// main is the program entry point. It sets the exit code if there is any error.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the program logic, including:
// - loading configuration
// - initializing logger
// - connecting to the database
// - creating the API server
// - handling shutdown signals
func run() error {
	rand.Seed(globaltime.Now().UnixNano())

	// Load Configuration
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Initialize Logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Info("application initializing")

	// Initialize Database
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename)
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite DB: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = dbconn.Close()
	}()

	db, err := database.New(dbconn)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}

	// Create API Server
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: db,
	})
	if err != nil {
		logger.WithError(err).Error("error creating API server instance")
		return fmt.Errorf("creating API server instance: %w", err)
	}

	// Register Web UI (if needed)
	router := apirouter.Handler()
	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}

	// Apply CORS
	router = applyCORSHandler(router)

	// Start API Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiserver := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%s", port),
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Run API Server in a separate goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Infof("API server listening on %s", apiserver.Addr)
		serverErrors <- apiserver.ListenAndServe()
	}()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("signal %v received, shutting down", sig)

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := apiserver.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("error during graceful shutdown")
			return fmt.Errorf("shutdown error: %w", err)
		}

		logger.Info("server stopped gracefully")
	}

	return nil
}
