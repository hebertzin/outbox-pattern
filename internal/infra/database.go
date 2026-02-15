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
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	DB       *sql.DB
}

func NewDatabaseConnection(
	host string,
	port int,
	user string,
	password string,
	dbname string,
) *DatabaseConnection {
	return &DatabaseConnection{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}

func (d *DatabaseConnection) Connect() error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.DBName,
	)

	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(MaxConnectionsOpen)
	if err := db.Ping(); err != nil {
		return err
	}

	d.DB = db

	return nil
}
