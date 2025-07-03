-- name: CreateTagStandard :one
INSERT INTO tag_standards (name, description, cloud_provider, version, content)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTagStandard :one
SELECT * FROM tag_standards
WHERE id = ?;

-- name: GetTagStandardByName :one
SELECT * FROM tag_standards
WHERE name = ?;

-- name: ListTagStandards :many
SELECT * FROM tag_standards
ORDER BY created_at DESC;

-- name: ListTagStandardsByProvider :many
SELECT * FROM tag_standards
WHERE cloud_provider = ?
ORDER BY created_at DESC;

-- name: UpdateTagStandard :one
UPDATE tag_standards
SET name = ?, description = ?, cloud_provider = ?, version = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTagStandard :exec
DELETE FROM tag_standards
WHERE id = ?;