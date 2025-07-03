-- name: CreateOperationLog :one
INSERT INTO operation_logs (operation_id, level, message, details)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetOperationLogs :many
SELECT * FROM operation_logs
WHERE operation_id = ?
ORDER BY created_at ASC;

-- name: GetOperationLogsByLevel :many
SELECT * FROM operation_logs
WHERE operation_id = ? AND level = ?
ORDER BY created_at ASC;

-- name: DeleteOperationLogs :exec
DELETE FROM operation_logs
WHERE operation_id = ?;