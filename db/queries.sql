-- name: CreateThread :one
INSERT INTO threads (title)
VALUES ($1)
RETURNING *;

-- name: GetThread :one
SELECT * FROM threads
WHERE id = $1;

-- name: ListThreads :many
SELECT * FROM threads
ORDER BY updated_at DESC;

-- name: UpdateThreadTitle :exec
UPDATE threads
SET title = $2, updated_at = now()
WHERE id = $1;

-- name: DeleteThread :exec
DELETE FROM threads
WHERE id = $1;

-- name: CreateMessage :one
INSERT INTO messages (thread_id, role, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMessagesByThread :many
SELECT * FROM messages
WHERE thread_id = $1
ORDER BY created_at ASC;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;
