package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yashtiwari22/email-dispatcher/backend/db"
)

// CreateCampaignRequest defines incoming JSON payload for campaign creation.
type CreateCampaignRequest struct {
	Title        string `json:"title"`
	Subject      string `json:"subject"`
	BodyTemplate string `json:"body_template"`
}

// UpdateStatusRequest defines incoming JSON payload for status updates.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// registerCampaignRoutes registers campaign HTTP handlers.
func (s *Server) registerCampaignRoutes() {
	s.router.HandleFunc("POST /api/v1/campaigns", s.handleCreateCampaign)
	s.router.HandleFunc("GET /api/v1/campaigns", s.handleListCampaigns)
	s.router.HandleFunc("GET /api/v1/campaigns/{id}", s.handleGetCampaign)
	s.router.HandleFunc("PATCH /api/v1/campaigns/{id}/status", s.handleUpdateCampaignStatus)
}

// handleCreateCampaign creates a new campaign record in database.
func (s *Server) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.BodyTemplate) == "" {
		respondError(w, http.StatusUnprocessableEntity, "title, subject, and body_template are required fields")
		return
	}

	campaign := db.Campaign{
		Title:        req.Title,
		Subject:      req.Subject,
		BodyTemplate: req.BodyTemplate,
		Status:       db.CampaignStatusDraft,
	}

	if err := s.db.Create(&campaign).Error; err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create campaign record")
		return
	}

	respondJSON(w, http.StatusCreated, campaign)
}

// handleListCampaigns returns all campaigns ordered by creation date.
func (s *Server) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	var campaigns []db.Campaign
	if err := s.db.Order("created_at desc").Find(&campaigns).Error; err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch campaigns")
		return
	}

	respondJSON(w, http.StatusOK, campaigns)
}

// handleGetCampaign fetches single campaign by path param ID.
func (s *Server) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var campaign db.Campaign
	if err := s.db.Preload("Recipients").First(&campaign, id).Error; err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	respondJSON(w, http.StatusOK, campaign)
}

// handleUpdateCampaignStatus updates status of a campaign.
func (s *Server) handleUpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	allowedStatuses := map[string]bool{
		db.CampaignStatusDraft:      true,
		db.CampaignStatusQueued:     true,
		db.CampaignStatusProcessing: true,
		db.CampaignStatusPaused:     true,
		db.CampaignStatusCompleted:  true,
		db.CampaignStatusFailed:     true,
	}

	if !allowedStatuses[req.Status] {
		respondError(w, http.StatusBadRequest, "Invalid campaign status value")
		return
	}

	if err := s.db.Model(&db.Campaign{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update campaign status")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}
