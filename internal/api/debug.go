package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleDebugNotion(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	err := s.notion.Ping(ctx)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
	})
}

func (s *Server) handleDebugSearchDatabases(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	dbs, err := s.notion.SearchDatabases(ctx)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}

	type liteDB struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	out := struct {
		Count int      `json:"count"`
		DBs   []liteDB `json:"dbs"`
	}{}

	for _, db := range dbs {
		title := ""
		if len(db.Title) > 0 {
			title = db.Title[0].PlainText
		}
		out.DBs = append(out.DBs, liteDB{
			ID:    db.ID,
			Title: title,
		})
	}
	out.Count = len(out.DBs)

	_ = json.NewEncoder(w).Encode(out)
}
