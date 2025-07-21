package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"

	"github.com/justcgh9/vk-internship-application/pkg/tracing"

	"github.com/justcgh9/vk-internship-application/internal/config"
	authhandler "github.com/justcgh9/vk-internship-application/internal/http/handlers/auth"
	listingshandler "github.com/justcgh9/vk-internship-application/internal/http/handlers/listings"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/justcgh9/vk-internship-application/internal/service/listing"
	"github.com/justcgh9/vk-internship-application/internal/storage/postgres"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
	"github.com/justcgh9/vk-internship-application/pkg/metrics"
)

func main() {

	cfg := config.MustLoad()

	logger.Init(slog.LevelDebug)
	logger.Log.Info("Starting application...")

	metrics.Init()

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		logger.Log.Error("Failed to connect to DB", slog.Any("err", err))
		os.Exit(1)
	}
	defer dbpool.Close()

	store := postgres.NewStorage(dbpool)

	authSvc := auth.New(store, auth.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL))
	listingSvc := listing.New(store)

	ctx := context.Background()
	shutdownTracing, err := tracing.Init(ctx, "vk-intern-app")
	if err != nil {
		logger.Log.Error("failed to initialize tracing", slog.Any("err", err))
		return
	}
	defer func() {
		if err := shutdownTracing(ctx); err != nil {
			logger.Log.Error("error shutting down tracer", slog.Any("err", err))
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: slog.NewLogLogger(logger.Log.Handler(), slog.LevelDebug)}))
	r.Use(middleware.Recoverer)
	r.Use(metrics.Middleware)
	r.Use(otelchi.Middleware("vk-intern-app"))

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	validate := validator.New()

	authHandler := authhandler.New(
		authSvc,
		validate,
	)

	r.Mount("/auth", authHandler.Routes(authSvc))

	listingsHandler := listingshandler.New(
		authSvc,
		listingSvc,
		validate,
	)

	r.Mount("/listings", listingsHandler.Routes(authSvc))

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		WriteTimeout: cfg.Server.Timeout,
	}

	go func() {
		logger.Log.Info("HTTP server listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Failed graceful shutdown", slog.Any("err", err))
	}
}
