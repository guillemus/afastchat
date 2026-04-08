package db

import (
	"errors"
	"fmt"

	"zombiezen.com/go/sqlite/sqlitex"
)

var ErrNotFound = errors.New("not found")

type Conn struct {
	pool *sqlitex.Pool
}

func Connect(dbPath string) (*Conn, error) {
	pool, err := sqlitex.NewPool(dbPath, sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("open db pool: %w", err)
	}

	return &Conn{pool: pool}, nil
}

func (c *Conn) Close() error {
	return c.pool.Close()
}
