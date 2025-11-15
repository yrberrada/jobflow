package store

import (
	"context"
	"database/sql"
)

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store { return &Store{DB: db} }

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS jobs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	external_id TEXT UNIQUE,
	title TEXT,
	company TEXT,
	location TEXT,
	url TEXT,
	work_mode TEXT,
	salary TEXT,
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS applications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	job_id INTEGER NOT NULL,
	status TEXT,
	outcome TEXT,
	applied_on TIMESTAMP NULL,
	interview_time TIMESTAMP NULL,
	notes TEXT,
	notion_page_id TEXT,
	FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS contacts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	job_id INTEGER NOT NULL,
	name TEXT,
	email TEXT,
	role TEXT,
	notes TEXT,
	FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
);
`)
	return err
}
