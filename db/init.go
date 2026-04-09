package db

import (
	"context"
	"errors"
	"fmt"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

var ErrNotFound = errors.New("not found")

type Pool struct {
	pool *sqlitex.Pool
}

func (x *Pool) WithConn(ctx context.Context, fn func(*sqlite.Conn) error) error {
	conn, err := x.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer x.pool.Put(conn)

	return fn(conn)
}

func Connect(dbPath string) (*Pool, error) {
	pool, err := sqlitex.NewPool(dbPath, sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("open db pool: %w", err)
	}

	return &Pool{pool: pool}, nil
}

func (x *Pool) Close() error {
	return x.pool.Close()
}
