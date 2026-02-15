package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"transaction-service/internal/infra"
	"transaction-service/internal/presentation"
	"transaction-service/internal/usecase"

	"github.com/gorilla/mux"
)

func main() {
	db := connectDatabase()
	defer db.Close()

	router := registerTransactionRoutes(db)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func connectDatabase() *sql.DB {
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("invalid DB_PORT: %v", err)
	}

	dbConn := infra.NewDatabaseConnection(
		host,
		port,
		user,
		password,
		dbName,
	)

	if err := dbConn.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return dbConn.DB
}

func registerTransactionRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()
	transactionHandler := transactionFactory(db)
	transactionHandler.RegisterRoutes(router)
	return router
}

func transactionFactory(db *sql.DB) *presentation.CreateTransactionHandler {
	r := infra.NewTransactionRepository(db)
	u := usecase.NewCreateTransactionUseCase(r)
	h := presentation.NewCreateTransactionHandler(u)

	return h
}
