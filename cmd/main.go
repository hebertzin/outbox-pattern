package main

import (
	"database/sql"
	"transaction-service/internal/infra"
	"transaction-service/internal/presentation"
	"transaction-service/internal/usecase"

	"github.com/gorilla/mux"
)

func main() {

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
