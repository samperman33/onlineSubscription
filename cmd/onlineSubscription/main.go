package main

import (
	"context"
	"log/slog"
	"net/http"
	"onlineSubscription/internal/config"
	"onlineSubscription/internal/handlers"
	"onlineSubscription/internal/storage"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.LoadCfg()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	log.Info("Initializing server", slog.String("address", cfg.HTTPServer.Address))
	//log.Info("config loaded", "addr", cfg.HTTPServer.Address)
	log.Debug("Logger debug mode enabled")

	//TODO: db

	db, err := storage.NewPostgresStorage(cfg)
	if err != nil {
		log.Error("Failed to connect db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	//TODO: routing
	router := chi.NewRouter()

	router.Get("/subscriptions", handlers.ListSubscriptionsHandler(db))
	router.Get("/subscriptions/aggregate", handlers.AggregateHandler(db))
	router.Post("/subscriptions", handlers.CreateSubscriptionHandler(db))
	router.Get("/subscriptions/{id}", handlers.GetSubscriptionHandler(db))
	router.Put("/subscriptions/{id}", handlers.UpdateSubscriptionHandler(db))
	router.Delete("/subscriptions/{id}", handlers.DeleteSubscriptionHandler(db))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("Starting server...", "addr", cfg.HTTPServer.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "err", err)
		}
	}()

	//TODO: shutdowning server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown err", "err", err)
	}
	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
