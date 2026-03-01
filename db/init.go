package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*Queries, error) {
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return New(conn), nil
}
