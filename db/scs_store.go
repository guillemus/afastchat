package db

import (
	"context"
	"log"
	"time"

	"github.com/alexedwards/scs/v2"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

var _ scs.CtxStore = (*SCSStore)(nil)
var _ scs.IterableStore = (*SCSStore)(nil)

// SCSStore is a direct port of the official SCS SQLite store, adapted only
// from database/sql to this app's zombiezen/sqlite pool:
// https://github.com/alexedwards/scs/blob/209de6e426de9259665975ce16b91331d228f052/sqlite3store/sqlite3store.go
// Keep SQL and cleanup behavior aligned with upstream unless intentional.
type SCSStore struct {
	pool        *Pool
	stopCleanup chan bool
}

func NewSCSStore(pool *Pool) *SCSStore {
	return NewSCSStoreWithCleanupInterval(pool, 5*time.Minute)
}

func NewSCSStoreWithCleanupInterval(pool *Pool, cleanupInterval time.Duration) *SCSStore {
	store := &SCSStore{pool: pool}
	if cleanupInterval > 0 {
		store.stopCleanup = make(chan bool)
		go store.startCleanup(cleanupInterval)
	}
	return store
}

func (x *SCSStore) Find(token string) ([]byte, bool, error) {
	return x.FindCtx(context.Background(), token)
}

func (x *SCSStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	var data []byte
	found := false

	err := x.pool.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT data
			FROM sessions
			WHERE token = :token AND julianday('now') < expiry
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":token": token},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				data = make([]byte, stmt.GetLen("data"))
				stmt.GetBytes("data", data)
				found = true
				return nil
			},
		})
	})
	if err != nil {
		return nil, false, err
	}

	return data, found, nil
}

func (x *SCSStore) Commit(token string, b []byte, expiry time.Time) error {
	return x.CommitCtx(context.Background(), token, b, expiry)
}

func (x *SCSStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	return x.pool.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			REPLACE INTO sessions (token, data, expiry)
			VALUES (:token, :data, julianday(:expiry))
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":token":  token,
				":data":   b,
				":expiry": expiry.UTC().Format("2006-01-02T15:04:05.999"),
			},
		})
	})
}

func (x *SCSStore) Delete(token string) error {
	return x.DeleteCtx(context.Background(), token)
}

func (x *SCSStore) DeleteCtx(ctx context.Context, token string) error {
	return x.pool.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM sessions
			WHERE token = :token
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":token": token},
		})
	})
}

func (x *SCSStore) All() (map[string][]byte, error) {
	sessions := map[string][]byte{}

	err := x.pool.WithConn(context.Background(), func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT token, data
			FROM sessions
			WHERE julianday('now') < expiry
		`, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				data := make([]byte, stmt.GetLen("data"))
				stmt.GetBytes("data", data)
				sessions[stmt.GetText("token")] = data
				return nil
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (x *SCSStore) StopCleanup() {
	if x.stopCleanup != nil {
		x.stopCleanup <- true
	}
}

func (x *SCSStore) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := x.deleteExpired(context.Background())
			if err != nil {
				log.Println(err)
			}
		case <-x.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

func (x *SCSStore) deleteExpired(ctx context.Context) error {
	return x.pool.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM sessions
			WHERE expiry < julianday('now')
		`, nil)
	})
}
