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
	getSegment "segmentify/internal/httpserver/handlers/segments/get"
	createUser "segmentify/internal/httpserver/handlers/users/create"
	getUserSegments "segmentify/internal/httpserver/handlers/users/get"
	downloadUserSegmentsHistory "segmentify/internal/httpserver/handlers/users/gethistory"
	updateUserSegments "segmentify/internal/httpserver/handlers/users/update"
	mwLogger "segmentify/internal/httpserver/middleware/logger"
	"segmentify/internal/lib/logger/sl"
	"segmentify/internal/storage/postgres"

	_ "segmentify/docs"

	_ "github.com/swaggo/swag"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	envTest = "test"
	envDev  = "dev"
	envProd = "prod"
)

// @title			Segmentify
// @description	Dynamic user segmentation service
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting segmentify", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	ctx := context.Background()

	storage, err := postgres.New(ctx, cfg.PostgresURI)
	if err != nil {
		log.Error("failed to start storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	if err := storage.Init(ctx); err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(
		middleware.RequestID,
		middleware.Logger,
		mwLogger.New(log),
		middleware.Recoverer,
	)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Route("/segments", func(r chi.Router) {
		r.Post("/", createSegment.New(ctx, log, storage))
		r.Delete("/{slug}", deleteSegment.New(ctx, log, storage))
		r.Get("/{slug}", getSegment.New(ctx, log, storage))
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/", createUser.New(ctx, log, storage))
		r.Get("/{id}/segments", getUserSegments.New(ctx, log, storage))
		r.Get("/{id}/download-segments-history", downloadUserSegmentsHistory.New(ctx, log, storage))
		r.Patch("/{id}/segments", updateUserSegments.New(ctx, log, storage))

	})

	log.Info("starting server", slog.String("address", cfg.Address))

	doneServer := make(chan os.Signal, 1)
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

	go startScheduler(ctx, log, storage)

	<-doneServer
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	case envTest, envDev:
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

func startScheduler(ctx context.Context, log *slog.Logger, storage *postgres.Storage) {
	for {
		rowsAffected, err := storage.RemoveExpiredUsersSegments(ctx)
		if err != nil {
			log.Error("failed to run a RemoveExpiredUsersSegments job", sl.Err(err))
		} else {
			log.Info("RemoveExpiredUsersSegments job completed", slog.Int64("rowsAffected", rowsAffected))
		}
		time.Sleep(time.Hour)
	}
}
