package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"transaction-service/config"
	_ "transaction-service/docs"
	infradb "transaction-service/infra/db"
	"transaction-service/infra/repository"
	"transaction-service/internal/core/broker"
	"transaction-service/internal/core/handler"
	"transaction-service/internal/core/usecase"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title          Transaction Service API
// @version        1.0
// @description    Transaction service using the outbox pattern for reliable event publishing.
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

	txRepo := repository.NewTransactionRepository(db)

	createUC := usecase.NewCreateTransactionUseCase(txRepo)
	statusUC := usecase.NewGetTransactionStatusUseCase(txRepo)
	balanceUC := usecase.NewGetBalanceUseCase(txRepo)

	txHandler := handler.NewHandler(createUC, statusUC, balanceUC)

	mux := http.NewServeMux()
	txHandler.RegisterRoutes(mux)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: mux,
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

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	}
}
