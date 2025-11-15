package notion

import (
	"context"

	gnt "github.com/dstotijn/go-notion"

	"jobflow.local/internal/domain"
)

type Client struct {
	api        *gnt.Client
	databaseID string
}

func New(token, databaseID string) *Client {
	return &Client{
		api:        gnt.NewClient(token),
		databaseID: databaseID,
	}
}

// Ping just tries a tiny QueryDatabase to see if the DB is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.api.QueryDatabase(ctx, c.databaseID, &gnt.DatabaseQuery{
		PageSize: 1,
	})
	return err
}

// SearchDatabases is used by /debug/notion/search to list DBs.
func (c *Client) SearchDatabases(ctx context.Context) ([]gnt.Database, error) {
	resp, err := c.api.Search(ctx, &gnt.SearchOpts{
		Filter: &gnt.SearchFilter{
			Property: "object",
			Value:    "database",
		},
		PageSize: 20,
	})
	if err != nil {
		return nil, err
	}

	var dbs []gnt.Database
	for _, obj := range resp.Results {
		if db, ok := obj.(gnt.Database); ok {
			dbs = append(dbs, db)
		}
	}

	return dbs, nil
}

// helper: build a valid Notion rich_text slice from a plain string.
func richText(s string) []gnt.RichText {
	if s == "" {
		return nil
	}
	return []gnt.RichText{
		{
			Text: &gnt.Text{
				Content: s,
			},
		},
	}
}

func buildJobPageProperties(job domain.Job, app domain.Application) gnt.DatabasePageProperties {
	props := gnt.DatabasePageProperties{}

	// Position — Title (required title property)
	if job.Title != "" {
		props["Position"] = gnt.DatabasePageProperty{
			Title: richText(job.Title),
		}
	}

	// Company — Text (rich_text)
	if job.Company != "" {
		props["Company"] = gnt.DatabasePageProperty{
			RichText: richText(job.Company),
		}
	}

	// Job Posting — URL
	if job.URL != "" {
		props["Job Posting"] = gnt.DatabasePageProperty{
			URL: &job.URL,
		}
	}

	// Work Mode — Select
	if job.WorkMode != "" {
		props["Work Mode"] = gnt.DatabasePageProperty{
			Select: &gnt.SelectOptions{
				Name: job.WorkMode,
			},
		}
	}

	// location — Text (rich_text, lowercase name)
	if job.Location != "" {
		props["location"] = gnt.DatabasePageProperty{
			RichText: richText(job.Location),
		}
	}

	// Salary — Text (rich_text)
	if job.Salary != "" {
		props["Salary"] = gnt.DatabasePageProperty{
			RichText: richText(job.Salary),
		}
	}

	// Stage — Select
	if app.Stage != "" {
		props["Stage"] = gnt.DatabasePageProperty{
			Select: &gnt.SelectOptions{
				Name: app.Stage,
			},
		}
	}

	// Outcome — Select
	if app.Outcome != "" {
		props["Outcome"] = gnt.DatabasePageProperty{
			Select: &gnt.SelectOptions{
				Name: app.Outcome,
			},
		}
	}

	// Notes — Text (rich_text)
	if app.Notes != "" {
		props["Notes"] = gnt.DatabasePageProperty{
			RichText: richText(app.Notes),
		}
	}

	// Next Interview — Date (optional)
	if app.InterviewTime != nil {
		dt := gnt.NewDateTime(*app.InterviewTime, true)
		props["Next Interview"] = gnt.DatabasePageProperty{
			Date: &gnt.Date{
				Start: dt,
			},
		}
	}

	return props
}

// CreateJobPage: create a new row in the Job Tracker (beta) DB.
func (c *Client) CreateJobPage(ctx context.Context, job domain.Job, app domain.Application) (string, error) {
	props := buildJobPageProperties(job, app)

	params := gnt.CreatePageParams{
		ParentType:             gnt.ParentTypeDatabase,
		ParentID:               c.databaseID,
		DatabasePageProperties: &props,
	}

	page, err := c.api.CreatePage(ctx, params)
	if err != nil {
		return "", err
	}
	return page.ID, nil
}
