-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ?;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY uploaded_at DESC;

-- name: CreateSession :one
INSERT INTO sessions (
    id, name, date, start_time, uploaded_at
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateSession :one
UPDATE sessions
SET name = ?, date = ?, start_time = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;

-- name: GetLatestSessionID :one
SELECT id FROM sessions
ORDER BY uploaded_at DESC
LIMIT 1;