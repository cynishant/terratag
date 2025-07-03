-- name: CreateOperation :one
INSERT INTO operations (type, status, standard_id, directory_path, filter_pattern, skip_pattern, settings)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetOperation :one
SELECT * FROM operations
WHERE id = ?;

-- name: ListOperations :many
SELECT * FROM operations
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListOperationsByType :many
SELECT * FROM operations
WHERE type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateOperationStatus :one
UPDATE operations
SET status = ?
WHERE id = ?
RETURNING *;

-- name: UpdateOperationStarted :one
UPDATE operations
SET status = ?, started_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateOperationCompleted :one
UPDATE operations
SET status = ?, completed_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteOperation :exec
DELETE FROM operations
WHERE id = ?;