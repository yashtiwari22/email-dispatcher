package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yashtiwari22/email-dispatcher/backend/db"
	"github.com/yashtiwari22/email-dispatcher/backend/engine"
)

// registerDLQRoutes registers DLQ management HTTP handlers.
func (s *Server) registerDLQRoutes() {
	s.router.HandleFunc("GET /api/v1/dlq", s.handleListDLQ)
	s.router.HandleFunc("POST /api/v1/dlq/{id}/replay", s.handleReplayDLQ)
	s.router.HandleFunc("DELETE /api/v1/dlq", s.handleClearDLQ)
}

// handleListDLQ returns list of dead-letter records filtered by optional status query param.
func (s *Server) handleListDLQ(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")

	query := s.db.Order("failed_at desc")
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var records []db.DLQRecord
	if err := query.Find(&records).Error; err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch DLQ records")
		return
	}

	respondJSON(w, http.StatusOK, records)
}

// handleReplayDLQ re-enqueues a failed job payload back into Asynq queue.
func (s *Server) handleReplayDLQ(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid DLQ record ID")
		return
	}

	var dlqRecord db.DLQRecord
	if err := s.db.First(&dlqRecord, uint(id)).Error; err != nil {
		respondError(w, http.StatusNotFound, "DLQ record not found")
		return
	}

	var payload engine.EmailTaskPayload
	if err := json.Unmarshal([]byte(dlqRecord.PayloadJSON), &payload); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "Failed to unmarshal DLQ payload JSON")
		return
	}

	if s.asynqClient != nil {
		task, err := engine.NewEmailDispatchTask(payload)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to construct dispatch task")
			return
		}
		if _, err := s.asynqClient.Enqueue(task); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to re-enqueue task into Redis queue")
			return
		}
	}

	// Mark DLQ record status as replayed
	s.db.Model(&dlqRecord).Update("status", db.DLQStatusReplayed)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":      dlqRecord.ID,
		"status":  db.DLQStatusReplayed,
		"message": "Job successfully replayed and enqueued",
	})
}

// handleClearDLQ deletes non-pending (replayed or discarded) DLQ records.
func (s *Server) handleClearDLQ(w http.ResponseWriter, r *http.Request) {
	result := s.db.Where("status != ?", db.DLQStatusPending).Delete(&db.DLQRecord{})
	if result.Error != nil {
		respondError(w, http.StatusInternalServerError, "Failed to purge DLQ records")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted_count": result.RowsAffected,
	})
}
