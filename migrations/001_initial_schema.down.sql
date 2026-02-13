-- Down migration: drop all tables created in 001_initial_schema.up.sql

DROP INDEX IF EXISTS idx_thermoworks_data_timestamp;
DROP INDEX IF EXISTS idx_thermoworks_data_session_id;
DROP INDEX IF EXISTS idx_events_session_id;
DROP INDEX IF EXISTS idx_stages_session_id;
DROP INDEX IF EXISTS idx_probes_session_id;

DROP TABLE IF EXISTS thermoworks_data;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS probes;
DROP TABLE IF EXISTS sessions;
