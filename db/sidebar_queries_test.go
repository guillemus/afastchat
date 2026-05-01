package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, pool *Pool) string {
	t.Helper()
	user, err := pool.CreateUser(context.Background(), "test@example.com", "hash")
	require.NoError(t, err)
	return user.ID
}

func TestGetSidebarTreeNestedFoldersDoesNotDependOnAlphabeticalOrder(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	parentID, err := pool.CreateFolder(ctx, userID, nil, "Z Parent")
	require.NoError(t, err)
	childID, err := pool.CreateFolder(ctx, userID, &parentID, "A Child")
	require.NoError(t, err)
	thread, err := pool.CreateThread(ctx, userID, "nested chat")
	require.NoError(t, err)
	require.NoError(t, pool.MoveThreadToFolder(ctx, userID, thread.ID, &childID))

	tree, err := pool.GetSidebarTree(ctx, userID, "")
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	require.Equal(t, parentID, tree.Folders[0].Folder.ID)
	require.Len(t, tree.Folders[0].Folders, 1)
	require.Equal(t, childID, tree.Folders[0].Folders[0].Folder.ID)
	require.Len(t, tree.Folders[0].Folders[0].Threads, 1)
	require.Equal(t, thread.ID, tree.Folders[0].Folders[0].Threads[0].ID)
	require.Equal(t, 2, tree.Folders[0].DeleteFolderCount)
	require.Equal(t, 1, tree.Folders[0].DeleteThreadCount)
}
