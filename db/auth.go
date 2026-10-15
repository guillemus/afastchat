package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type AuthProvider string

const AuthProviderGoogle AuthProvider = "google"

type LoginWithIdentityParams struct {
	Provider    AuthProvider
	Subject     string
	Email       string
	DisplayName string
}

func (x *Pool) LoginWithIdentity(ctx context.Context, arg LoginWithIdentityParams) (user User, err error) {
	err = x.WithConn(ctx, func(conn *sqlite.Conn) (err error) {
		defer sqlitex.Save(conn)(&err)

		found, err := scanOneUser(conn, `
			SELECT u.id, u.email, u.display_name, u.selected_model, u.created_at, u.updated_at
			FROM auth_identities i
			JOIN users u ON u.id = i.user_id
			WHERE i.provider = :provider AND i.subject = :subject
		`, map[string]any{":provider": string(arg.Provider), ":subject": arg.Subject}, &user)
		if err != nil {
			return fmt.Errorf("lookup auth identity: %w", err)
		}

		if found {
			if _, err := scanOneUser(conn, `
				UPDATE users
				SET email = :email, display_name = :display_name, updated_at = CURRENT_TIMESTAMP
				WHERE id = :id
				RETURNING id, email, display_name, selected_model, created_at, updated_at
			`, map[string]any{":email": arg.Email, ":display_name": arg.DisplayName, ":id": user.ID}, &user); err != nil {
				return fmt.Errorf("refresh authenticated user: %w", err)
			}
			if err := sqlitex.Execute(conn, `
				UPDATE auth_identities
				SET updated_at = CURRENT_TIMESTAMP
				WHERE provider = :provider AND subject = :subject
			`, &sqlitex.ExecOptions{Named: map[string]any{":provider": string(arg.Provider), ":subject": arg.Subject}}); err != nil {
				return fmt.Errorf("update auth identity: %w", err)
			}
			return nil
		}

		found, err = scanOneUser(conn, `
			SELECT id, email, display_name, selected_model, created_at, updated_at
			FROM users
			WHERE lower(email) = lower(:email)
		`, map[string]any{":email": arg.Email}, &user)
		if err != nil {
			return fmt.Errorf("lookup user by verified email: %w", err)
		}

		if !found {
			if _, err := scanOneUser(conn, `
				INSERT INTO users (id, email, display_name)
				VALUES (:id, :email, :display_name)
				RETURNING id, email, display_name, selected_model, created_at, updated_at
			`, map[string]any{":id": uuid.New().String(), ":email": arg.Email, ":display_name": arg.DisplayName}, &user); err != nil {
				return fmt.Errorf("insert user: %w", err)
			}
		}

		if err := sqlitex.Execute(conn, `
			INSERT INTO auth_identities (id, user_id, provider, subject)
			VALUES (:id, :user_id, :provider, :subject)
		`, &sqlitex.ExecOptions{Named: map[string]any{
			":id":       uuid.New().String(),
			":user_id":  user.ID,
			":provider": string(arg.Provider),
			":subject":  arg.Subject,
		}}); err != nil {
			return fmt.Errorf("insert auth identity: %w", err)
		}
		return nil
	})
	return user, err
}

func (x *Pool) DeleteUser(ctx context.Context, userID string) ([]string, error) {
	var objectKeys []string
	err := x.WithConn(ctx, func(conn *sqlite.Conn) (err error) {
		defer sqlitex.Save(conn)(&err)

		named := map[string]any{":user_id": userID}
		objectKeys, err = selectAttachmentKeys(conn, `
			SELECT object_key FROM message_attachments WHERE user_id = :user_id
		`, named)
		if err != nil {
			return err
		}

		for _, query := range []string{
			`DELETE FROM usage WHERE user_id = :user_id`,
			`DELETE FROM message_attachments WHERE user_id = :user_id`,
			`DELETE FROM messages WHERE thread_id IN (SELECT id FROM threads WHERE user_id = :user_id)`,
			`DELETE FROM threads WHERE user_id = :user_id`,
			`DELETE FROM openrouter_connections WHERE user_id = :user_id`,
			`DELETE FROM user_favorite_models WHERE user_id = :user_id`,
			`DELETE FROM auth_identities WHERE user_id = :user_id`,
			`DELETE FROM users WHERE id = :user_id`,
		} {
			if err := sqlitex.Execute(conn, query, &sqlitex.ExecOptions{Named: named}); err != nil {
				return fmt.Errorf("delete user: %w", err)
			}
		}
		return nil
	})
	return objectKeys, err
}

func scanOneUser(conn *sqlite.Conn, query string, named map[string]any, user *User) (bool, error) {
	found := false
	err := sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
		Named: named,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			found = true
			*user = scanUser(stmt)
			return nil
		},
	})
	return found, err
}
