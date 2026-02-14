-- SQLite schema for twchart application
-- Initial schema migration

-- Main sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    date DATETIME NOT NULL,
    start_time DATETIME,
    uploaded_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Probes table (one-to-many with sessions)
CREATE TABLE IF NOT EXISTS probes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Stages table (one-to-many with sessions)
CREATE TABLE IF NOT EXISTS stages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    start DATETIME NOT NULL,
    end DATETIME,
    duration INTEGER, -- stored as nanoseconds
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Events table (one-to-many with sessions)
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    note TEXT NOT NULL,
    time DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Thermoworks data table (one-to-many with sessions)
CREATE TABLE IF NOT EXISTS thermoworks_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    probe1_temp REAL,
    probe2_temp REAL,
    probe3_temp REAL,
    probe4_temp REAL,
    probe5_temp REAL,
    probe6_temp REAL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_probes_session_id ON probes(session_id);
CREATE INDEX IF NOT EXISTS idx_stages_session_id ON stages(session_id);
CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
CREATE INDEX IF NOT EXISTS idx_thermoworks_data_session_id ON thermoworks_data(session_id);
CREATE INDEX IF NOT EXISTS idx_thermoworks_data_timestamp ON thermoworks_data(timestamp);
