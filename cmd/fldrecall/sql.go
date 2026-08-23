package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Global or package-level DB reference for the requested helper function
var dbConn *sql.DB

// CreateTables initializes the schema for the program.
func CreateTables(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	queries := []string{
		// Snapshots Table
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL
		);`,

		// Unique Directories Table
		`CREATE TABLE IF NOT EXISTS directories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE
		);`,

		// Many-to-Many Junction Table
		`CREATE TABLE IF NOT EXISTS snapshot_directories (
			snapshot_id INTEGER NOT NULL,
			directory_id INTEGER NOT NULL,
			PRIMARY KEY (snapshot_id, directory_id),
			FOREIGN KEY (snapshot_id) REFERENCES snapshots(id) ON DELETE CASCADE,
			FOREIGN KEY (directory_id) REFERENCES directories(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindDirectoryId searches the database and returns the ID of a directory that matches the given path.
// Returns -1 and nil error if the directory does not exist.
func FindDirectoryId(db *sql.DB, path string) (int, error) {
	var id int
	query := `SELECT id FROM directories WHERE path = ?`
	err := db.QueryRow(query, path).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil // Not found
		}
		return -1, err // Real database error
	}
	return id, nil
}

// SaveSnapshot saves a complete Snapshot struct (including its Directories) to the database.
// Existing entries (snapshot or directories) are resolves and new entries are inserted into the database.
func SaveSnapshot(db *sql.DB, snapshot *Snapshot) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Create a copy of the snapshot which may get modified with new ids
	copy := *snapshot

	// Resolve or insert the Snapshot ID
	if copy.Id == -1 {
		result, err := tx.Exec(`INSERT INTO snapshots (timestamp) VALUES (?)`, copy.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to insert snapshot: %w", err)
		}
		snapID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get snapshot id: %w", err)
		}
		copy.Id = int(snapID)
	}

	// Prepare statements inside transaction for speed
	insertDirStmt, err := tx.Prepare(`INSERT INTO directories (path) VALUES (?)`)
	if err != nil {
		return err
	}
	defer insertDirStmt.Close()

	// Loop through directories and resolve ids
	for i, dir := range copy.Directories {
		// If ID is -1, check the DB to know if we must insert or query
		if dir.Id == -1 {

			// Query via transaction space to see context of current block
			existingId, err := FindDirectoryId(db, dir.Path)
			if existingId == -1 {
				// Truly new path; insert it into the DB
				res, err := insertDirStmt.Exec(dir.Path)
				if err != nil {
					return fmt.Errorf("failed to insert new path %s: %w", dir.Path, err)
				}
				newId, err := res.LastInsertId()
				if err != nil {
					return err
				}
				copy.Directories[i].Id = int(newId)
			} else if err != nil {
				// An actual error occured, abort
				return err
			} else {
				// Found an existing path; reuse its ID
				copy.Directories[i].Id = existingId
			}
		}
	}

	// Prepare statements inside transaction for speed
	bridgeStmt, err := tx.Prepare(`INSERT OR IGNORE INTO snapshot_directories (snapshot_id, directory_id) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer bridgeStmt.Close()

	// Loop through directory ids and link them with a snapshot id
	for _, dir := range copy.Directories {
		_, err = bridgeStmt.Exec(copy.Id, dir.Id)
		if err != nil {
			return fmt.Errorf("failed to link snapshot to directory: %w", err)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit the transaction: %w", err)
	}

	// Update the given snapshot
	*snapshot = copy

	return nil
}

// InsertSnapshots inserts multiple Snapshots in the database
func InsertSnapshots(db *sql.DB, snapshots []*Snapshot) error {
	for _, snapshot := range snapshots {
		err := SaveSnapshot(db, snapshot)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindSnapshotId searches for a snapshot by its exact timestamp.
// Returns -1 and nil error if no snapshot matches the timestamp.
func FindSnapshotId(db *sql.DB, timestampd time.Time) (int, error) {
	var id int
	query := `SELECT id FROM snapshots WHERE timestamp = ?`
	err := db.QueryRow(query, timestampd).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil
		}
		return -1, err
	}
	return id, nil
}

// LoadSnapshot loads a complete Snapshot struct (including its Directories) using its ID.
func LoadSnapshot(db *sql.DB, id int) (*Snapshot, error) {
	// Fetch the main snapshot metadata
	var snapshot Snapshot
	querySnap := `SELECT id, timestamp FROM snapshots WHERE id = ?`
	err := db.QueryRow(querySnap, id).Scan(&snapshot.Id, &snapshot.Timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("snapshot with id %d not found: %w)", id, err)
		}
		return nil, err
	}

	// Fetch all directories associated with this snapshot via the junction table
	queryDirs := `
		SELECT d.id, d.path 
		FROM directories d
		JOIN snapshot_directories ON d.id = snapshot_directories.directory_id
		WHERE snapshot_directories.snapshot_id = ?`

	rows, err := db.Query(queryDirs, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dir Directory
		err := rows.Scan(&dir.Id, &dir.Path)
		if err != nil {
			return nil, err
		}
		snapshot.Directories = append(snapshot.Directories, dir)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// GetLatestSnapshotId finds the Id of the most recent snapshot in the database.
func GetLatestSnapshotId(db *sql.DB) (int, error) {
	latestId := InvalidId

	// Order by timestamp descending and limit to 1 to find the newest entry
	query := `SELECT id FROM snapshots ORDER BY timestamp DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&latestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvalidId, fmt.Errorf("no snapshots found in database")
		}
		return InvalidId, err
	}

	return latestId, nil
}

// GetLatestSnapshot finds the most recent snapshot in the database and loads it completely.
func GetLatestSnapshot(db *sql.DB) (*Snapshot, error) {
	latestId, err := GetLatestSnapshotId(db)
	if err != nil {
		return nil, err
	}

	snapshot, err := LoadSnapshot(db, latestId)
	return snapshot, err
}

// GetLastNSnapshotIds finds the Ids of the most recent snapshots in the database, limitting results to a maximum of `max` results.
func GetLastNSnapshotIds(db *sql.DB, max int) ([]int, error) {
	latestIds := []int{}

	// Order by timestamp descending and limit to 1 to find the newest entry
	query := `SELECT id FROM snapshots ORDER BY timestamp DESC LIMIT ?`
	rows, err := db.Query(query, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		latestIds = append(latestIds, id)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return latestIds, nil
}

// GetLastNSnapshots finds the most recent snapshots in the database, limitting results to a maximum of `max` results.
func GetLastNSnapshots(db *sql.DB, max int) ([]*Snapshot, error) {
	latestIds, err := GetLastNSnapshotIds(db, max)
	if err != nil {
		return nil, err
	}

	snapshots := []*Snapshot{}

	for _, id := range latestIds {
		snapshot, err := LoadSnapshot(db, id)
		if err != nil {
			return nil, err
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}
