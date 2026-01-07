-- name: GetEventsBySession :many
SELECT * FROM events
WHERE session_id = ?
ORDER BY time;

-- name: CreateEvent :one
INSERT INTO events (session_id, note, time)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteEventsBySession :exec
DELETE FROM events WHERE session_id = ?;