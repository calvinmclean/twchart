-- name: GetStagesBySession :many
SELECT * FROM stages
WHERE session_id = ?
ORDER BY start;

-- name: CreateStage :one
INSERT INTO stages (session_id, name, start, end, duration)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateStage :one
UPDATE stages
SET end = ?, duration = ?
WHERE id = ?
RETURNING *;

-- name: DeleteStagesBySession :exec
DELETE FROM stages WHERE session_id = ?;