package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Server holds application dependencies for HTTP request handlers.
type Server struct {
	db          *gorm.DB
	asynqClient *asynq.Client
	router      *http.ServeMux
}

// NewServer initializes a new API Server with configured routes.
func NewServer(database *gorm.DB, client *asynq.Client) *Server {
	s := &Server{
		db:          database,
		asynqClient: client,
		router:      http.NewServeMux(),
	}

	s.routes()
	return s
}

// ServeHTTP implements http.Handler with CORS and structured logging middleware.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for frontend web dashboard
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("[API] %s %s", r.Method, r.URL.Path)
	s.router.ServeHTTP(w, r)
}

// respondJSON helper writes a JSON response with status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("[API Error] Failed to write JSON response: %v", err)
		}
	}
}

// respondError helper writes a JSON error object.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
