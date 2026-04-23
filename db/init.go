package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed schema.sql
var schema string

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("duplicate email")

type Pool struct {
	pool *sqlitex.Pool
}

func NewPool(pool *sqlitex.Pool) *Pool {
	return &Pool{pool: pool}
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
		PoolSize:    10,
		PrepareConn: prepareConn,
	})
	if err != nil {
		return nil, fmt.Errorf("open db pool: %w", err)
	}

	return NewPool(pool), nil
}

func prepareConn(conn *sqlite.Conn) error {
	if err := sqlitex.Execute(conn, "PRAGMA journal_mode = WAL;", nil); err != nil {
		return err
	}
	if err := sqlitex.Execute(conn, "PRAGMA busy_timeout = 5000;", nil); err != nil {
		return err
	}
	if err := sqlitex.Execute(conn, "PRAGMA synchronous = NORMAL;", nil); err != nil {
		return err
	}
	if err := sqlitex.Execute(conn, "PRAGMA cache_size = -262144;", nil); err != nil {
		return err
	}
	if err := sqlitex.Execute(conn, "PRAGMA foreign_keys = ON;", nil); err != nil {
		return err
	}
	if err := sqlitex.Execute(conn, "PRAGMA temp_store = MEMORY;", nil); err != nil {
		return err
	}
	return nil
}

func (x *Pool) Close() error {
	return x.pool.Close()
}
