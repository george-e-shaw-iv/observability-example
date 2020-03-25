package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/george-e-shaw-iv/observability-example/cmd/listd/handlers"
	"github.com/george-e-shaw-iv/observability-example/internal/platform/config"
	"github.com/george-e-shaw-iv/observability-example/internal/platform/db"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.FromEnvironment()
	if err != nil {
		fmt.Printf("error getting config from environment: %v", err)
		os.Exit(1)
	}

	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.LogLevel))
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zCfg.Build()
	if err != nil {
		fmt.Printf("error initializing logger: %v\n", err)
		os.Exit(1)
	}

	dbCfg := db.Config{
		User: cfg.DBUser,
		Pass: cfg.DBPass,
		Name: cfg.DBName,
		Host: cfg.DBHost,
		Port: cfg.DBPort,
	}
	dbc, err := db.NewConnection(dbCfg, logger)
	if err != nil {
		logger.Error("connect to database", zap.Error(err))
		os.Exit(1)
	}

	server := http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.DaemonPort),
		Handler:        handlers.NewApplication(dbc, logger),
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Start listening for requests made to the daemon and create a channel
	// to collect non-HTTP related server errors on.
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server started", zap.String("address", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown of the daemon.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Waiting for an osSignal or a non-HTTP related server error.
	select {
	case e := <-serverErrors:
		logger.Error("server failed to start", zap.Error(e))

	case <-osSignals:
	}

	// Gracefully shutdown server once an exit signal or error is received.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	var shutdownFailure bool
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown of server failed",
			zap.Duration("timeout", cfg.ShutdownTimeout),
			zap.Error(err))

		logger.Info("attempting to forcefully close server")
		if err := server.Close(); err != nil {
			logger.Error("failed to forcefully close server", zap.Error(err))
		}

		shutdownFailure = true
	}

	// Release resources from the shutdown context.
	cancel()

	// Close database connection before program exits.
	if err := dbc.Close(); err != nil {
		logger.Error("close database", zap.Error(err))
	}

	// A shutdown failure had occurred, exit with the proper status code.
	if shutdownFailure {
		os.Exit(1)
	}

	// Exit with a successful status code.
	os.Exit(0)
}
