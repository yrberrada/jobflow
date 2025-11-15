package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// OpenSQLite opens a SQLite DB and enables foreign keys.
func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
