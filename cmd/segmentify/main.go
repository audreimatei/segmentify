package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"segmentify/internal/config"
	createSegment "segmentify/internal/httpserver/handlers/segments/create"
	deleteSegment "segmentify/internal/httpserver/handlers/segments/delete"
	getSegmentBySlug "segmentify/internal/httpserver/handlers/segments/getbyslug"
	createUser "segmentify/internal/httpserver/handlers/users/create"
	getUserSegments "segmentify/internal/httpserver/handlers/users/get"
	downloadUserSegmentsHistory "segmentify/internal/httpserver/handlers/users/gethistory"
	updateUserSegments "segmentify/internal/httpserver/handlers/users/update"
	mwLogger "segmentify/internal/httpserver/middleware/logger"
	"segmentify/internal/lib/logger/sl"
	"segmentify/internal/storage/postgres"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting segmentify", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(cfg.PostgresURI)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	storage.Init()

	router := chi.NewRouter()

	router.Use(
		middleware.RequestID,
		middleware.Logger,
		mwLogger.New(log),
		middleware.Recoverer,
	)

	router.Route("/segments", func(r chi.Router) {
		r.Post("/", createSegment.New(log, storage))
		r.Get("/{slug}", getSegmentBySlug.New(log, storage))
		r.Delete("/", deleteSegment.New(log, storage))
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/", createUser.New(log, storage))
		r.Get("/{userID}/segments", getUserSegments.New(log, storage))
		r.Get("/{userID}/download-segments-history", downloadUserSegmentsHistory.New(log, storage))
		r.Patch("/{userID}/segments", updateUserSegments.New(log, storage))

	})

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
