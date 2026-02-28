package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

var (
	MaxConnectionsOpen int = 25;
)
type Connection struct {
	Hostname string
	Username string
	Password string
	Database string
	Port     int
}

func NewConnection(hostname, username, password, database string, port int) *Connection {
	return &Connection{
		Hostname: hostname,
		Username: username,
		Password: password,
		Database: database,
		Port:     port,
	}
}

func (c *Connection) Connect(ctx context.Context) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Username,
		c.Password,
		c.Hostname,
		c.Port,
		c.Database,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(MaxConnectionsOpen)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
