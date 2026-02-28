package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"users-services/application/usecase"
	"users-services/config"
	"users-services/infra/db/postgres"
	"users-services/infra/db/repository"
	"users-services/infra/messaging"
	"users-services/infra/worker"
	"users-services/presentation"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn := postgres.NewConnection(
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
	)
	db, err := conn.Connect(ctx)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("connected to database")

	userRepo := repository.NewUserRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	createUserUC := usecase.NewCreateUserUseCase(userRepo)

	userHandler := presentation.NewUserHandler(createUserUC)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", userHandler.Create)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: mux,
	}

	publisher := messaging.NewLogEventPublisher(logger)
	outboxWorker := worker.NewOutboxWorker(outboxRepo, publisher, cfg.Worker.Interval, logger)

	go outboxWorker.Run(ctx)

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
