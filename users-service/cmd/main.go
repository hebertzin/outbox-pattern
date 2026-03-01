package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"users-service/config"
	"users-service/infra/broker"
	infradb "users-service/infra/db"
	"users-service/infra/repository"
	"users-service/internal/core/handler"
	"users-service/internal/core/usecase"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// @title          Users Service API
// @version        1.0
// @description    Users service using the outbox pattern for reliable event publishing.
// @host           localhost:8080
// @BasePath       /
// @schemes        http
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := infradb.Connect(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to database")

	rabbit := broker.NewRabbitMQ(cfg.RabbitMQ.URL)
	if err := rabbit.Connect(); err != nil {
		logger.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer rabbit.Close()
	logger.Info("connected to rabbitmq")

	userRepo := repository.NewUserRepository(db)
	userHandler := handler.NewHandlerFactory(usecase.NewFactory(userRepo, logger))

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      handler.MetricsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("starting HTTP server", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	}
}
