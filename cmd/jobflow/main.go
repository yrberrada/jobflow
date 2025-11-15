package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"jobflow.local/internal/api"
	ncli "jobflow.local/internal/notion"
	"jobflow.local/internal/store"
)

func mask(s string) string {
	if len(s) <= 10 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-4:]
}

// normalizeNotionID removes dashes if present.
func normalizeNotionID(id string) string {
	id = strings.TrimSpace(id)
	return strings.ReplaceAll(id, "-", "")
}

func main() {
	_ = godotenv.Load()

	rawNotionToken := os.Getenv("NOTION_TOKEN")
	rawNotionDBID := os.Getenv("NOTION_DB_ID")
	sqlitePath := os.Getenv("JOBFLOW_DB")
	port := os.Getenv("PORT")

	if port == "" {
		// You’re already using 8081, keep that.
		port = "8081"
	}
	if sqlitePath == "" {
		sqlitePath = "jobflow.sqlite"
	}
	if rawNotionToken == "" || rawNotionDBID == "" {
		log.Fatal("NOTION_TOKEN and NOTION_DB_ID must be set in your environment (.env)")
	}

	notionDBID := normalizeNotionID(rawNotionDBID)

	log.Println("=== JobFlow Startup Sanity ===")
	log.Println("Using Notion DB ID (raw):    ", rawNotionDBID)
	log.Println("Using Notion DB ID (norm):   ", notionDBID)
	log.Println("Using Notion Token (masked): ", mask(rawNotionToken))
	log.Println("SQLite file:                  ", sqlitePath)
	log.Println("HTTP port:                    ", port)
	log.Println("==============================")

	// SQLite
	db, err := store.OpenSQLite(sqlitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	st := store.New(db)
	if err := st.Migrate(context.Background()); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("SQLite ready at:", sqlitePath)

	// Notion client + ping
	nc := ncli.New(rawNotionToken, notionDBID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := nc.Ping(ctx); err != nil {
		log.Fatalf("Notion ping failed: %v", err)
	}
	log.Println("Notion connection OK.")

	// HTTP API
	s := api.New(st, nc)
	addr := ":" + port
	log.Println("HTTP listening on", addr)
	if err := s.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
