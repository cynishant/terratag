-- name: CreateOperationResult :one
INSERT INTO operation_results (operation_id, file_path, resource_type, resource_name, line_number, snippet, action, violation_type, details)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetOperationResults :many
SELECT * FROM operation_results
WHERE operation_id = ?
ORDER BY created_at ASC;

-- name: GetOperationResultsByAction :many
SELECT * FROM operation_results
WHERE operation_id = ? AND action = ?
ORDER BY created_at ASC;

-- name: CountOperationResultsByAction :one
SELECT COUNT(*) FROM operation_results
WHERE operation_id = ? AND action = ?;

-- name: DeleteOperationResults :exec
DELETE FROM operation_results
WHERE operation_id = ?;