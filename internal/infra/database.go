package infra

import (
	"database/sql"
	"fmt"
)

var (
	MaxConnectionsOpen = 25
	DriverName         = "postgres"
)

type DatabaseConnection struct {
	DB *sql.DB
}

func NewDatabaseConnection(
	host string,
	port int,
	user string,
	password string,
	dbname string,
) (*DatabaseConnection, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbname,
	)

	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(MaxConnectionsOpen)
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DatabaseConnection{
		DB: db,
	}, nil
}
