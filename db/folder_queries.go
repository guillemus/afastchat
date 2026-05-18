package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func cleanFolderName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}

func nullableTextParam(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func (x *Pool) CreateFolder(ctx context.Context, userID string, parentID *string, name string) (string, error) {
	name = cleanFolderName(name)
	if name == "" {
		return "", ErrNotFound
	}

	id := uuid.New().String()
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO folders (id, user_id, parent_id, name)
			VALUES (:id, :user_id, :parent_id, :name)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id, ":user_id": userID, ":parent_id": nullableTextParam(parentID), ":name": name},
		})
		if err != nil {
			return fmt.Errorf("insert folder: %w", err)
		}
		return nil
	})
	return id, err
}

func (x *Pool) RenameFolder(ctx context.Context, userID, folderID, name string) error {
	name = cleanFolderName(name)
	if name == "" {
		return ErrNotFound
	}
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			UPDATE folders
			SET name = :name, updated_at = datetime('now')
			WHERE id = :id AND user_id = :user_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":name": name, ":id": folderID, ":user_id": userID},
		})
		if err != nil {
			return fmt.Errorf("rename folder: %w", err)
		}
		if conn.Changes() == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (x *Pool) MoveThreadToFolder(ctx context.Context, userID, threadID string, folderID *string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			UPDATE threads
			SET folder_id = :folder_id, updated_at = datetime('now')
			WHERE id = :id
				AND user_id = :user_id
				AND (:folder_id IS NULL OR :folder_id IN (SELECT id FROM folders WHERE user_id = :user_id))
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":folder_id": nullableTextParam(folderID), ":id": threadID, ":user_id": userID},
		})
		if err != nil {
			return fmt.Errorf("move thread: %w", err)
		}
		if conn.Changes() == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (x *Pool) MoveFolder(ctx context.Context, userID, folderID string, parentID *string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			WITH RECURSIVE descendants(id) AS (
				SELECT id FROM folders WHERE parent_id = :id AND user_id = :user_id
				UNION ALL
				SELECT folders.id FROM folders JOIN descendants ON folders.parent_id = descendants.id
			)
			UPDATE folders
			SET parent_id = :parent_id, updated_at = datetime('now')
			WHERE id = :id
				AND user_id = :user_id
				AND (:parent_id IS NULL OR :parent_id != :id)
				AND (:parent_id IS NULL OR :parent_id IN (SELECT id FROM folders WHERE user_id = :user_id))
				AND (:parent_id IS NULL OR :parent_id NOT IN (SELECT id FROM descendants))
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":parent_id": nullableTextParam(parentID), ":id": folderID, ":user_id": userID},
		})
		if err != nil {
			return fmt.Errorf("move folder: %w", err)
		}
		if conn.Changes() == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (x *Pool) DeleteFolderRecursive(ctx context.Context, userID, folderID string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		end, err := sqlitex.ImmediateTransaction(conn)
		if err != nil {
			return err
		}
		defer end(&err)

		if err = sqlitex.Execute(conn, `
			WITH RECURSIVE subtree(id) AS (
				SELECT id FROM folders WHERE id = :id AND user_id = :user_id
				UNION ALL
				SELECT folders.id FROM folders JOIN subtree ON folders.parent_id = subtree.id
			)
			DELETE FROM threads WHERE user_id = :user_id AND folder_id IN (SELECT id FROM subtree)
		`, &sqlitex.ExecOptions{Named: map[string]any{":id": folderID, ":user_id": userID}}); err != nil {
			return fmt.Errorf("delete threads: %w", err)
		}
		if err = sqlitex.Execute(conn, `
			WITH RECURSIVE subtree(id) AS (
				SELECT id FROM folders WHERE id = :id AND user_id = :user_id
				UNION ALL
				SELECT folders.id FROM folders JOIN subtree ON folders.parent_id = subtree.id
			)
			DELETE FROM folders WHERE user_id = :user_id AND id IN (SELECT id FROM subtree)
		`, &sqlitex.ExecOptions{Named: map[string]any{":id": folderID, ":user_id": userID}}); err != nil {
			return fmt.Errorf("delete folders: %w", err)
		}
		if conn.Changes() == 0 {
			return ErrNotFound
		}
		return nil
	})
}
