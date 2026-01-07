-- name: GetProbesBySession :many
SELECT * FROM probes
WHERE session_id = ?;

-- name: CreateProbe :one
INSERT INTO probes (session_id, name, position)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteProbesBySession :exec
DELETE FROM probes WHERE session_id = ?;