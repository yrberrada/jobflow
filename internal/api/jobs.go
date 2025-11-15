package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"jobflow.local/internal/domain"
)

// JSON payload we expect from the browser / requests.http.
type applyRequest struct {
	ExternalID    string  `json:"external_id"`
	Position      string  `json:"position"`
	Company       string  `json:"company"`
	Location      string  `json:"location"`
	URL           string  `json:"url"`
	WorkMode      string  `json:"work_mode"`
	Salary        string  `json:"salary"`
	Notes         string  `json:"notes"`
	Stage         string  `json:"stage"`
	Outcome       string  `json:"outcome"`
	NextInterview *string `json:"next_interview"` // ISO8601 (RFC3339), optional
}

// handleApply is the main entry point for recording an application.
// 1) Upsert Job + Application in SQLite
// 2) Create a row in Notion
// 3) Save the Notion page id back into the DB (best effort)
func (s *Server) handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req applyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[/apply] JSON decode error: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[/apply] incoming payload: %+v", req)

	// Map request -> domain.Job
	job := domain.Job{
		ExternalID: req.ExternalID,
		Title:      req.Position,
		Company:    req.Company,
		Location:   req.Location,
		URL:        req.URL,
		WorkMode:   req.WorkMode,
		Salary:     req.Salary,
	}

	// Parse next interview (optional)
	var interviewTime *time.Time
	if req.NextInterview != nil && *req.NextInterview != "" {
		t, err := time.Parse(time.RFC3339, *req.NextInterview)
		if err != nil {
			log.Printf("[/apply] bad next_interview format (expected RFC3339) %q: %v", *req.NextInterview, err)
			http.Error(w, "invalid next_interview datetime (expected RFC3339)", http.StatusBadRequest)
			return
		}
		interviewTime = &t
	}

	// Map request -> domain.Application
	app := domain.Application{
		Stage:         req.Stage,
		Outcome:       req.Outcome,
		Notes:         req.Notes,
		InterviewTime: interviewTime,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// 1) Upsert locally (job + application)
	if err := s.store.UpsertJobAndApplication(ctx, &job, &app); err != nil {
		log.Printf("[/apply] DB error in UpsertJobAndApplication: %v", err)
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[/apply] DB upsert ok: job_id=%d application_id=%d", job.ID, app.ID)

	// 2) Create row in Notion
	pageID, err := s.notion.CreateJobPage(ctx, job, app)
	if err != nil {
		log.Printf("[/apply] Notion error in CreateJobPage: %v", err)
		http.Error(w, "notion error: "+err.Error(), http.StatusBadGateway)
		return
	}
	log.Printf("[/apply] Notion page created: %s", pageID)

	// 3) Save Notion page id back into DB (best-effort; if it fails we log but still return 201)
	if err := s.store.SaveNotionPageID(ctx, app.ID, pageID); err != nil {
		log.Printf("[/apply] warning: SaveNotionPageID failed: %v", err)
	}

	// Success response
	resp := map[string]any{
		"ok":             true,
		"job_id":         job.ID,
		"application_id": app.ID,
		"notion_page_id": pageID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[/apply] encode response error: %v", err)
	}
}
