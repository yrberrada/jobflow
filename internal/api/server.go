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
	s.mux.HandleFunc("POST /apply", s.handleApply)
}

func (s *Server) Handle(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Listen(addr string) error {
	log.Println("Server starting…")
	return http.ListenAndServe(addr, s.mux)
}
