package db

import (
	"context"
	"fmt"
	"strings"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type SidebarTree struct {
	Folders []SidebarFolder
	Threads []Thread
}

type SidebarFolder struct {
	Folder            Folder
	Folders           []SidebarFolder
	Threads           []Thread
	DeleteFolderCount int
	DeleteThreadCount int
}

type FolderDeleteImpact struct {
	FolderCount int
	ThreadCount int
}

func (x *Pool) GetFolderDeleteImpact(ctx context.Context, userID, folderID string) (FolderDeleteImpact, error) {
	var out FolderDeleteImpact
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			WITH RECURSIVE subtree(id) AS (
				SELECT id FROM folders WHERE id = :id AND user_id = :user_id
				UNION ALL
				SELECT folders.id FROM folders JOIN subtree ON folders.parent_id = subtree.id
			)
			SELECT
				(SELECT count(*) FROM subtree) AS folder_count,
				(SELECT count(*) FROM threads WHERE user_id = :user_id AND folder_id IN (SELECT id FROM subtree)) AS thread_count
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": folderID, ":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				out.FolderCount = int(stmt.GetInt64("folder_count"))
				out.ThreadCount = int(stmt.GetInt64("thread_count"))
				return nil
			},
		})
	})
	if err != nil {
		return out, err
	}
	if out.FolderCount == 0 {
		return out, ErrNotFound
	}
	return out, nil
}

func (x *Pool) GetSidebarTree(ctx context.Context, userID, query string) (SidebarTree, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	folders, err := x.listSidebarFolders(ctx, userID)
	if err != nil {
		return SidebarTree{}, err
	}
	threads, err := x.listSidebarThreads(ctx, userID, query)
	if err != nil {
		return SidebarTree{}, err
	}
	threadCounts, err := x.listSidebarThreadCounts(ctx, userID)
	if err != nil {
		return SidebarTree{}, err
	}

	foldersByID := make(map[string]Folder, len(folders))
	visibleThreadsByFolderID := make(map[string][]Thread, len(folders))
	childIDsByFolderID := make(map[string][]string, len(folders))
	var rootFolderIDs []string
	for _, folder := range folders {
		foldersByID[folder.ID] = folder
		if folder.ParentID == nil {
			rootFolderIDs = append(rootFolderIDs, folder.ID)
		} else {
			childIDsByFolderID[*folder.ParentID] = append(childIDsByFolderID[*folder.ParentID], folder.ID)
		}
	}

	var tree SidebarTree
	for _, thread := range threads {
		if thread.FolderID == nil {
			tree.Threads = append(tree.Threads, thread)
			continue
		}
		if _, ok := foldersByID[*thread.FolderID]; ok {
			visibleThreadsByFolderID[*thread.FolderID] = append(visibleThreadsByFolderID[*thread.FolderID], thread)
		}
	}

	var buildFolder func(id string) (SidebarFolder, bool)
	buildFolder = func(id string) (SidebarFolder, bool) {
		node := SidebarFolder{
			Folder:            foldersByID[id],
			Threads:           visibleThreadsByFolderID[id],
			DeleteFolderCount: 1,
			DeleteThreadCount: threadCounts[id],
		}
		for _, childID := range childIDsByFolderID[id] {
			child, visible := buildFolder(childID)
			node.DeleteFolderCount += child.DeleteFolderCount
			node.DeleteThreadCount += child.DeleteThreadCount
			if visible {
				node.Folders = append(node.Folders, child)
			}
		}
		return node, query == "" || len(node.Threads) > 0 || len(node.Folders) > 0
	}
	for _, folderID := range rootFolderIDs {
		folder, visible := buildFolder(folderID)
		if visible {
			tree.Folders = append(tree.Folders, folder)
		}
	}

	return tree, nil
}

func (x *Pool) listSidebarFolders(ctx context.Context, userID string) ([]Folder, error) {
	var folders []Folder
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT id, user_id, parent_id, name, created_at, updated_at
			FROM folders
			WHERE user_id = :user_id
			ORDER BY lower(name) ASC, name ASC
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				folders = append(folders, scanFolder(stmt))
				return nil
			},
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	return folders, nil
}

func (x *Pool) listSidebarThreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	counts := map[string]int{}
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT folder_id, count(*) AS thread_count
			FROM threads
			WHERE user_id = :user_id AND folder_id IS NOT NULL
			GROUP BY folder_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				counts[stmt.GetText("folder_id")] = int(stmt.GetInt64("thread_count"))
				return nil
			},
		})
	})
	if err != nil {
		return nil, fmt.Errorf("count threads: %w", err)
	}
	return counts, nil
}

func (x *Pool) listSidebarThreads(ctx context.Context, userID, query string) ([]Thread, error) {
	var threads []Thread
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		sql := `
			SELECT id, user_id, folder_id, title, archived_at, is_streaming, created_at, updated_at
			FROM threads
			WHERE user_id = :user_id AND archived_at IS NULL
		`
		named := map[string]any{":user_id": userID}
		if query != "" {
			sql += ` AND lower(title) LIKE :query`
			named[":query"] = "%" + query + "%"
		}
		sql += ` ORDER BY created_at DESC`

		return sqlitex.Execute(conn, sql, &sqlitex.ExecOptions{
			Named: named,
			ResultFunc: func(stmt *sqlite.Stmt) error {
				threads = append(threads, scanThread(stmt))
				return nil
			},
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	return threads, nil
}
