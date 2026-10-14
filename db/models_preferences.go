package db

import (
	"context"

	"zombiezen.com/go/sqlite"
)

func (x *Pool) UpdateUserSelectedModel(ctx context.Context, userID, modelID string) error {
	return x.exec(ctx, "update user selected model", `
		UPDATE users
		SET selected_model = :model_id, updated_at = CURRENT_TIMESTAMP
		WHERE id = :user_id
	`, map[string]any{":user_id": userID, ":model_id": modelID})
}

func (x *Pool) ListFavoriteModels(ctx context.Context, userID string) ([]string, error) {
	return queryAll(ctx, x, "query favorite models", `
		SELECT model_id
		FROM user_favorite_models
		WHERE user_id = :user_id
		ORDER BY created_at, model_id
	`, map[string]any{":user_id": userID}, func(stmt *sqlite.Stmt) (string, error) {
		return stmt.GetText("model_id"), nil
	})
}

func (x *Pool) SetFavoriteModel(ctx context.Context, userID, modelID string, favorite bool) error {
	query := `
		DELETE FROM user_favorite_models
		WHERE user_id = :user_id AND model_id = :model_id
	`
	if favorite {
		query = `
			INSERT INTO user_favorite_models (user_id, model_id)
			VALUES (:user_id, :model_id)
			ON CONFLICT (user_id, model_id) DO NOTHING
		`
	}
	return x.exec(ctx, "set favorite model", query, map[string]any{":user_id": userID, ":model_id": modelID})
}
