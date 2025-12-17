/*
Webapi boots the primary HTTP server for the project. It wires configuration,
logging, database connectivity, CORS, and the embedded web UI (when built
with the `webui` tag) into a single executable.

Usage:

webapi [flags]

Configuration values are loaded via loadConfiguration() from flags,
environment variables, or the optional YAML file path.

Exit codes:

0  graceful shutdown (signal or expected stop)
>0 failure during startup, runtime, or shutdown

The program migrates the embedded database schema to the latest version on
startup before accepting requests.
*/
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

// main delegates to run() and exits with a non-zero status if initialization or
// shutdown fail. Keeping the logic in run() simplifies testing and future reuse.
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// run orchestrates the lifecycle of the API server:
//   - read configuration values and apply defaults
//   - initialize logging
//   - connect to external dependencies (SQLite in this build)
//   - construct the API router and wrap it with the web UI + CORS handlers
//   - start the HTTP server and monitor for shutdown signals or fatal errors
//   - gracefully shut down the server and connected resources
func run() error {
	rand.Seed(globaltime.Now().UnixNano())
	// Load Configuration and defaults
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// Init logging
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Infof("application initializing")

	// if err := os.MkdirAll("data", 0755); err != nil {
	// 	logger.WithError(err).Error("error creating data directory")
	// 	return fmt.Errorf("creating data directory: %w", err)
	// }

	// Start Database
	logger.Println("initializing database support")
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename)
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
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

	// Start (main) API server
	logger.Info("initializing API server")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Create the API router
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: db,
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}

	// Apply CORS policy
	router = applyCORSHandler(router)

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	go func() {
		logger.Infof("API listening on %s", apiserver.Addr)
		serverErrors <- apiserver.ListenAndServe()
		logger.Infof("stopping API server")
	}()

	// Waiting for shutdown signal or POSIX signals
	select {
	case err := <-serverErrors:
		// Non-recoverable server error
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("signal %v received, start shutdown", sig)

		// Asking API server to shut down and load shed.
		err := apirouter.Close()
		if err != nil {
			logger.WithError(err).Warning("graceful shutdown of apirouter error")
		}

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and load shed.
		err = apiserver.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Warning("error during graceful shutdown of HTTP server")
			err = apiserver.Close()
		}

		// Log the status of this shutdown.
		// switch {
		// case sig == syscall.SIGSTOP:
		// 	return errors.New("integrity issue caused shutdown")
		// case err != nil:
		// 	return fmt.Errorf("could not stop server gracefully: %w", err)
		// }

		//// a fix because I'm using windows
		switch {
		case sig == syscall.SIGTERM || sig == os.Interrupt:
			return errors.New("received shutdown signal")
		case err != nil:
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
