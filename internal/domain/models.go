package domain

import "time"

type Job struct {
	ID          int64
	ExternalID  string
	Title       string
	Company     string
	Location    string
	URL         string
	WorkMode    string
	Salary      string
	Description string
	CreatedAt   time.Time
}

type Application struct {
	ID            int64
	JobID         int64
	Stage         string
	Outcome       string
	AppliedOn     *time.Time
	InterviewTime *time.Time
	Notes         string
	NotionPageID  string
}
