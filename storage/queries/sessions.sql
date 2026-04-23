-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ?;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY uploaded_at DESC
LIMIT ?
OFFSET ?;

-- name: ListSessionsByType :many
SELECT * FROM sessions
WHERE type = ?
ORDER BY uploaded_at DESC
LIMIT ?
OFFSET ?;

-- name: CountSessions :one
SELECT COUNT(*) FROM sessions;

-- name: CountSessionsByType :one
SELECT COUNT(*) FROM sessions WHERE type = ?;

-- name: CreateSession :one
INSERT INTO sessions (
    id, name, type, date, start_time, uploaded_at
) VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateSession :one
UPDATE sessions
SET name = ?, type = ?, date = ?, start_time = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;

-- name: GetLatestSessionID :one
SELECT id FROM sessions
ORDER BY uploaded_at DESC
LIMIT 1;
