-- name: GetThermoworksDataBySession :many
SELECT * FROM thermoworks_data
WHERE session_id = ?
ORDER BY timestamp;

-- name: CreateThermoworksData :one
INSERT INTO thermoworks_data (
    session_id, timestamp, probe1_temp, probe2_temp, probe3_temp, probe4_temp, probe5_temp, probe6_temp
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: DeleteThermoworksDataBySession :exec
DELETE FROM thermoworks_data WHERE session_id = ?;