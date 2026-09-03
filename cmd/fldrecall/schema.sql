-- Snapshots Table
CREATE TABLE IF NOT EXISTS snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL
);

-- Unique Directories Table
CREATE TABLE IF NOT EXISTS directories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE
);

-- Many-to-Many Junction Table
CREATE TABLE IF NOT EXISTS snapshot_directories (
    snapshot_id INTEGER NOT NULL,
    directory_id INTEGER NOT NULL,
    PRIMARY KEY (snapshot_id, directory_id),
    FOREIGN KEY (snapshot_id) REFERENCES snapshots(id) ON DELETE CASCADE,
    FOREIGN KEY (directory_id) REFERENCES directories(id) ON DELETE CASCADE
);