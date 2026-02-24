package repository

import (
	"database/sql"
	"fmt"
)

var (
	MaxConnectionsOpen = 25
	DriverName         = "postgres"
)

type (
	DatabaseConnection struct {
		host     string
		port     int
		user     string
		password string
		dbName   string
		db       *sql.DB
	}
)

func NewDatabaseConnection(
	host string,
	port int,
	user string,
	password string,
	dbname string,
) *DatabaseConnection {
	return &DatabaseConnection{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbName:   dbname,
	}
}

func (d *DatabaseConnection) Connect() error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.host,
		d.port,
		d.user,
		d.password,
		d.dbName,
	)

	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(MaxConnectionsOpen)
	if err := db.Ping(); err != nil {
		return err
	}

	d.db = db

	return nil
}
