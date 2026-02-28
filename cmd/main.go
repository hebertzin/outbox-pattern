package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"transaction-service/internal/core/broker"
	"transaction-service/internal/core/httphandler/messagehandler"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Transaction Service API
// @version         1.0
// @description     API for creating transactions using Outbox Pattern.
// @BasePath        /api/v1
func main() {
	rabbitMq := broker.NewRabbitMQ("")

	rabbitMq.Connect()

	serveMux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	registerTransactionRoutes(serveMux, rabbitMq)

	go func() {
		log.Println("server running on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := srv.Close(); err != nil {
			log.Printf("server close error: %v", err)
		}
	}

	log.Println("server stopped")
}

func registerSwagger(router *mux.Router) {
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}

func registerTransactionRoutes(mux *http.ServeMux, b *broker.RabbitMQ) {
	handler := messagehandler.NewTransactionMessageHandler(b)

	handler.RegisterRoutes(mux)
}
