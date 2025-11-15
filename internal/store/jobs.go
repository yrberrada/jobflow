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

	// Upsert job
	if job.ExternalID != "" {
		var existingID int64
		row := tx.QueryRowContext(ctx, `SELECT id FROM jobs WHERE external_id = ?`, job.ExternalID)
		switch err := row.Scan(&existingID); err {
		case sql.ErrNoRows:
			// Insert new
			res, err := tx.ExecContext(ctx, `
				INSERT INTO jobs (external_id, title, company, location, url, work_mode, salary, description)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				job.ExternalID, job.Title, job.Company, job.Location,
				job.URL, job.WorkMode, job.Salary, job.Description,
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
			// Update existing
			job.ID = existingID
			_, err := tx.ExecContext(ctx, `
				UPDATE jobs
				SET title = ?, company = ?, location = ?, url = ?, work_mode = ?, salary = ?, description = ?
				WHERE id = ?`,
				job.Title, job.Company, job.Location, job.URL, job.WorkMode, job.Salary, job.Description, job.ID,
			)
			if err != nil {
				return err
			}

		default:
			return err
		}
	} else {
		// No external ID, just insert
		res, err := tx.ExecContext(ctx, `
			INSERT INTO jobs (external_id, title, company, location, url, work_mode, salary, description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			job.ExternalID, job.Title, job.Company, job.Location,
			job.URL, job.WorkMode, job.Salary, job.Description,
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

	// Insert application
	res, err := tx.ExecContext(ctx, `
		INSERT INTO applications (job_id, status, outcome, applied_on, interview_time, notes, notion_page_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		job.ID,
		app.Stage, // stored in "status" column for now
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

	committed = true
	return tx.Commit()
}
