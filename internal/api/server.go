package api

import (
	"log"
	"net/http"

	"jobflow.local/internal/notion"
	"jobflow.local/internal/store"
)

type Server struct {
	store  *store.Store
	notion *notion.Client
	mux    *http.ServeMux
}

func New(st *store.Store, n *notion.Client) *Server {
	s := &Server{
		store:  st,
		notion: n,
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /debug/notion", s.handleDebugNotion)
	s.mux.HandleFunc("GET /debug/notion/search", s.handleDebugSearchDatabases)

	// CORS preflight + main handler
	s.mux.HandleFunc("OPTIONS /apply", s.handleApply) // same func handles OPTIONS shortcut
	s.mux.HandleFunc("POST /apply", s.handleApply)
}

// Helper used by handlers to allow browser extension → API calls.
func writeCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // later you can restrict to https://www.linkedin.com
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) Handle(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Listen(addr string) error {
	log.Println("Server starting…")
	return http.ListenAndServe(addr, s.mux)
}
