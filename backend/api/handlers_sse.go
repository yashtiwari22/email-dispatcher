package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yashtiwari22/email-dispatcher/backend/db"
)

// CampaignProgressEvent defines real-time JSON payload sent over SSE stream.
type CampaignProgressEvent struct {
	CampaignID      uint    `json:"campaign_id"`
	Status          string  `json:"status"`
	TotalRecipients int     `json:"total_recipients"`
	SentCount       int     `json:"sent_count"`
	FailedCount     int     `json:"failed_count"`
	ProgressPct     float64 `json:"progress_pct"`
}

// registerSSERoutes registers SSE streaming handlers.
func (s *Server) registerSSERoutes() {
	s.router.HandleFunc("GET /api/v1/campaigns/{id}/stream", s.handleCampaignSSEStream)
}

// handleCampaignSSEStream streams real-time delivery metrics to frontend clients.
func (s *Server) handleCampaignSSEStream(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}
	campaignID := uint(id)

	// Verify http.ResponseWriter supports flushing for SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Streaming unsupported by server")
		return
	}

	// Set required SSE HTTP response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Client closed connection
			return
		case <-ticker.C:
			var campaign db.Campaign
			if err := s.db.Select("id, status, total_recipients, sent_count, failed_count").First(&campaign, campaignID).Error; err != nil {
				// Send error event and close stream
				fmt.Fprintf(w, "event: error\ndata: {\"error\": \"campaign not found\"}\n\n")
				flusher.Flush()
				return
			}

			pct := 0.0
			if campaign.TotalRecipients > 0 {
				pct = (float64(campaign.SentCount+campaign.FailedCount) / float64(campaign.TotalRecipients)) * 100.0
			}

			event := CampaignProgressEvent{
				CampaignID:      campaign.ID,
				Status:          campaign.Status,
				TotalRecipients: campaign.TotalRecipients,
				SentCount:       campaign.SentCount,
				FailedCount:     campaign.FailedCount,
				ProgressPct:     pct,
			}

			eventBytes, err := json.Marshal(event)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", string(eventBytes))
				flusher.Flush()
			}

			// If campaign reached terminal state, send final event and close stream
			if campaign.Status == db.CampaignStatusCompleted || campaign.Status == db.CampaignStatusFailed {
				return
			}
		}
	}
}
