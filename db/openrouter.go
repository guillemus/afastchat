package db

import (
	"context"

	"zombiezen.com/go/sqlite"
)

func (x *Pool) GetOpenRouterSealedKey(ctx context.Context, userID string) ([]byte, error) {
	return queryOne(ctx, x, "query openrouter connection", `
		SELECT sealed_key
		FROM openrouter_connections
		WHERE user_id = :user_id
	`, map[string]any{":user_id": userID}, func(stmt *sqlite.Stmt) ([]byte, error) {
		sealed := make([]byte, stmt.ColumnLen(0))
		stmt.ColumnBytes(0, sealed)
		return sealed, nil
	})
}

func (x *Pool) SaveOpenRouterConnection(ctx context.Context, userID string, sealedKey []byte) error {
	return x.exec(ctx, "save openrouter connection", `
		INSERT INTO openrouter_connections (user_id, sealed_key)
		VALUES (:user_id, :sealed_key)
		ON CONFLICT (user_id) DO UPDATE
		SET sealed_key = excluded.sealed_key, updated_at = CURRENT_TIMESTAMP
	`, map[string]any{":user_id": userID, ":sealed_key": sealedKey})
}

func (x *Pool) DeleteOpenRouterConnection(ctx context.Context, userID string) error {
	return x.exec(ctx, "delete openrouter connection", `
		DELETE FROM openrouter_connections
		WHERE user_id = :user_id
	`, map[string]any{":user_id": userID})
}
