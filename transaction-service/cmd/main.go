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
	infradb "transaction-service/infra/db"
	"transaction-service/infra/repository"
	"transaction-service/internal/core/broker"
	"transaction-service/internal/core/handler"
	"transaction-service/internal/core/usecase"
)

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

	txHandler := handler.NewTransactionHandler(createUC, statusUC, balanceUC)

	mux := http.NewServeMux()
	txHandler.RegisterRoutes(mux)

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
