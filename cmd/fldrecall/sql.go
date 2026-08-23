package main

import (
	"database/sql"
	"errors"
	"fmt"

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

// SaveSnapshot saves the snapshot, resolves/inserts directories, and bridges them.
func SaveSnapshot(db *sql.DB, originalSnapshot *Snapshot) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Create a copy of the snapshot which may get modified with new ids
	copy := *originalSnapshot

	// Resolve or insert the Snapshot ID
	if copy.Id == -1 {
		result, err := tx.Exec(`INSERT INTO snapshots (timestamp) VALUES (?)`, copy.TimeStamp)
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
	*originalSnapshot = copy

	return nil
}
