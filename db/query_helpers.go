package db

import (
	"context"
	"fmt"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// exec runs a database-writing statement that returns no rows.
// Use for INSERT, UPDATE, DELETE, and other statements that mutate database state.
func (x *Pool) exec(ctx context.Context, label, query string, named map[string]any) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		if err := sqlitex.Execute(conn, query, &sqlitex.ExecOptions{Named: named}); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
		return nil
	})
}

// execReturning runs a database-writing statement that returns one row.
// Use for INSERT/UPDATE/DELETE ... RETURNING, never for read-only SELECT queries.
func execReturning[T any](ctx context.Context, x *Pool, label, query string, named map[string]any, scan func(*sqlite.Stmt) (T, error)) (T, error) {
	var out T
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		found := false
		err := sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
			Named: named,
			ResultFunc: func(stmt *sqlite.Stmt) error {
				found = true
				var err error
				out, err = scan(stmt)
				return err
			},
		})
		if err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
		if !found {
			return ErrNotFound
		}
		return nil
	})
	return out, err
}

// queryOne runs a read-only query that returns exactly one row.
// Use for SELECT queries; it returns ErrNotFound when the query returns no rows.
func queryOne[T any](ctx context.Context, x *Pool, label, query string, named map[string]any, scan func(*sqlite.Stmt) (T, error)) (T, error) {
	return execReturning(ctx, x, label, query, named, scan)
}

// queryAll runs a read-only query that returns zero or more rows.
// Use for SELECT queries; it never mutates database state.
func queryAll[T any](ctx context.Context, x *Pool, label, query string, named map[string]any, scan func(*sqlite.Stmt) (T, error)) ([]T, error) {
	var out []T
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
			Named: named,
			ResultFunc: func(stmt *sqlite.Stmt) error {
				row, err := scan(stmt)
				if err != nil {
					return err
				}
				out = append(out, row)
				return nil
			},
		})
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return out, nil
}
