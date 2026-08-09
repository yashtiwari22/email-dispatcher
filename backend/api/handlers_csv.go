package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/yashtiwari22/email-dispatcher/backend/db"
	engine "github.com/yashtiwari22/email-dispatcher/backend/engine"
	"gorm.io/gorm"
)

// CSVUploadResponse summarizes outcome of CSV stream parsing and job queuing.
type CSVUploadResponse struct {
	CampaignID   uint `json:"campaign_id"`
	TotalParsed  int  `json:"total_parsed"`
	TotalQueued  int  `json:"total_queued"`
	InvalidCount int  `json:"invalid_count"`
}

// registerCSVRoutes registers CSV upload routes.
func (s *Server) registerCSVRoutes() {
	s.router.HandleFunc("POST /api/v1/campaigns/upload", s.handleCSVUpload)
}

// handleCSVUpload handles chunked CSV parsing and instant batch job enqueuing.
func (s *Server) handleCSVUpload(w http.ResponseWriter, r *http.Request) {
	// 1. Limit upload body size to 32MB max
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	campaignIDStr := r.FormValue("campaign_id")
	campaignIDUint, err := strconv.ParseUint(campaignIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or missing campaign_id form parameter")
		return
	}
	campaignID := uint(campaignIDUint)

	var campaign db.Campaign
	if err := s.db.First(&campaign, campaignID).Error; err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Missing or invalid file field in multipart request")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		respondError(w, http.StatusBadRequest, "CSV file is empty or invalid header row")
		return
	}

	emailIdx, nameIdx := -1, -1
	for i, h := range headers {
		headerClean := strings.ToLower(strings.TrimSpace(h))
		if headerClean == "email" {
			emailIdx = i
		} else if headerClean == "name" {
			nameIdx = i
		}
	}

	if emailIdx == -1 {
		respondError(w, http.StatusBadRequest, "CSV header must contain an 'email' column")
		return
	}

	var totalParsed, totalQueued, invalidCount int
	batchSize := 250
	recipientBatch := make([]db.Recipient, 0, batchSize)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			invalidCount++
			continue
		}

		totalParsed++

		email := strings.TrimSpace(record[emailIdx])
		if email == "" || !strings.Contains(email, "@") {
			invalidCount++
			continue
		}

		name := ""
		if nameIdx != -1 && nameIdx < len(record) {
			name = strings.TrimSpace(record[nameIdx])
		}

		rec := db.Recipient{
			CampaignID: campaignID,
			Name:       name,
			Email:      email,
			Status:     db.RecipientStatusPending,
		}
		recipientBatch = append(recipientBatch, rec)

		if len(recipientBatch) >= batchSize {
			queued := s.flushRecipientBatch(campaign, recipientBatch)
			totalQueued += queued
			recipientBatch = recipientBatch[:0]
		}
	}

	// Flush remaining recipients
	if len(recipientBatch) > 0 {
		queued := s.flushRecipientBatch(campaign, recipientBatch)
		totalQueued += queued
	}

	// Update campaign totals & status
	s.db.Model(&campaign).Updates(map[string]interface{}{
		"total_recipients": gorm.Expr("total_recipients + ?", totalQueued),
		"status":           db.CampaignStatusQueued,
	})

	respondJSON(w, http.StatusOK, CSVUploadResponse{
		CampaignID:   campaignID,
		TotalParsed:  totalParsed,
		TotalQueued:  totalQueued,
		InvalidCount: invalidCount,
	})
}

// flushRecipientBatch saves recipients to DB and enqueues task payloads into Redis Asynq queue.
func (s *Server) flushRecipientBatch(campaign db.Campaign, recipients []db.Recipient) int {
	if len(recipients) == 0 {
		return 0
	}

	if err := s.db.CreateInBatches(recipients, len(recipients)).Error; err != nil {
		fmt.Printf("[CSV Error] Batch insert recipients failed: %v\n", err)
		return 0
	}

	queuedCount := 0
	for _, r := range recipients {
		if s.asynqClient != nil {
			payload := engine.EmailTaskPayload{
				CampaignID:     campaign.ID,
				RecipientID:    r.ID,
				RecipientName:  r.Name,
				RecipientEmail: r.Email,
				SubjectTmpl:    campaign.Subject,
				BodyTmpl:       campaign.BodyTemplate,
			}

			task, err := engine.NewEmailDispatchTask(payload)
			if err == nil {
				_, _ = s.asynqClient.Enqueue(task)
				queuedCount++
			}
		} else {
			queuedCount++
		}
	}

	return queuedCount
}
