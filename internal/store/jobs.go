package store

import (
	"context"
	"database/sql"

	"jobflow.local/internal/domain"
)

// UpsertJobAndApplication:
// - If ExternalID is present, update or insert the job
// - Always insert a new application row
func (s *Store) UpsertJobAndApplication(ctx context.Context, job *domain.Job, app *domain.Application) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// --- 1) Upsert job (by external_id if present) ---

	if job.ExternalID != "" {
		var existingID int64
		row := tx.QueryRowContext(ctx, `SELECT id FROM jobs WHERE external_id = ?`, job.ExternalID)

		switch err := row.Scan(&existingID); err {
		case sql.ErrNoRows:
			// Insert new job
			res, err := tx.ExecContext(ctx, `
				INSERT INTO jobs (external_id, title, company, location, url, work_mode, salary, description)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				job.ExternalID,
				job.Title,
				job.Company,
				job.Location,
				job.URL,
				job.WorkMode,
				job.Salary,
				job.Description,
			)
			if err != nil {
				return err
			}
			jobID, err := res.LastInsertId()
			if err != nil {
				return err
			}
			job.ID = jobID

		case nil:
			// Update existing job
			job.ID = existingID
			_, err := tx.ExecContext(ctx, `
				UPDATE jobs
				SET title = ?, company = ?, location = ?, url = ?, work_mode = ?, salary = ?, description = ?
				WHERE id = ?`,
				job.Title,
				job.Company,
				job.Location,
				job.URL,
				job.WorkMode,
				job.Salary,
				job.Description,
				job.ID,
			)
			if err != nil {
				return err
			}

		default:
			// Some other error from Scan
			return err
		}
	} else {
		// No external ID → always insert a fresh job
		res, err := tx.ExecContext(ctx, `
			INSERT INTO jobs (external_id, title, company, location, url, work_mode, salary, description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			job.ExternalID,
			job.Title,
			job.Company,
			job.Location,
			job.URL,
			job.WorkMode,
			job.Salary,
			job.Description,
		)
		if err != nil {
			return err
		}
		jobID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		job.ID = jobID
	}

	// --- 2) Insert Application row ---

	res, err := tx.ExecContext(ctx, `
		INSERT INTO applications (job_id, status, outcome, applied_on, interview_time, notes, notion_page_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		job.ID,
		// We store Stage into "status" column:
		app.Stage,
		app.Outcome,
		app.AppliedOn,
		app.InterviewTime,
		app.Notes,
		app.NotionPageID,
	)
	if err != nil {
		return err
	}
	appID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	app.ID = appID
	app.JobID = job.ID

	committed = true
	return tx.Commit()
}
