-- name: CreateThread :one
INSERT INTO threads (id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetThread :one
SELECT * FROM threads
WHERE id = $1;

-- name: ListThreads :many
SELECT * FROM threads
WHERE archived_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateThreadTitle :exec
UPDATE threads
SET title = $2, updated_at = now()
WHERE id = $1;

-- name: ArchiveThread :exec
UPDATE threads
SET archived_at = now(), updated_at = now()
WHERE id = $1;

-- name: DeleteThread :exec
DELETE FROM threads
WHERE id = $1;

-- name: CreateMessage :one
INSERT INTO messages (id, thread_id, role, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMessagesByThread :many
SELECT * FROM messages
WHERE thread_id = $1
ORDER BY created_at ASC;

-- name: UpdateMessageContent :exec
UPDATE messages
SET content = $2
WHERE id = $1;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;
