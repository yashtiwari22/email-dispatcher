package main

import (
	"net/http"
)

// routes maps HTTP routes to their respective handler functions.
func (s *Server) routes() {
	s.router.HandleFunc("GET /healthz", s.handleHealthz)
	s.router.HandleFunc("GET /readyz", s.handleReadyz)
	s.registerCampaignRoutes()
	s.registerCSVRoutes()
	s.registerSSERoutes()
	s.registerDLQRoutes()
}

// handleHealthz handles liveness probe.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleReadyz handles readiness probe by verifying DB connection.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "error",
				"db":     "unreachable",
			})
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"db":     "connected",
	})
}
