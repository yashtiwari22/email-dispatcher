package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yashtiwari22/email-dispatcher/backend/db"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) (*Server, *gorm.DB) {
	database, err := db.ConnectSQLite(":memory:")
	require.NoError(t, err)

	err = db.AutoMigrate(database)
	require.NoError(t, err)

	srv := NewServer(database, nil)
	return srv, database
}

func TestHealthEndpoints(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Test GET /healthz
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)

	// Test GET /readyz
	reqReady := httptest.NewRequest("GET", "/readyz", nil)
	recReady := httptest.NewRecorder()
	srv.ServeHTTP(recReady, reqReady)

	assert.Equal(t, http.StatusOK, recReady.Code)
	assert.Contains(t, recReady.Body.String(), `"status":"ready"`)
}

func TestCampaignLifecycle(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Create Campaign
	createBody := map[string]string{
		"title":         "Welcome Onboarding",
		"subject":       "Welcome to our Platform {{.Name}}",
		"body_template": "Hello {{.Name}}, your email is {{.Email}}.",
	}
	bodyBytes, _ := json.Marshal(createBody)

	req := httptest.NewRequest("POST", "/api/v1/campaigns", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdCampaign db.Campaign
	err := json.Unmarshal(rec.Body.Bytes(), &createdCampaign)
	require.NoError(t, err)
	assert.Equal(t, uint(1), createdCampaign.ID)
	assert.Equal(t, "Welcome Onboarding", createdCampaign.Title)

	// 2. List Campaigns
	reqList := httptest.NewRequest("GET", "/api/v1/campaigns", nil)
	recList := httptest.NewRecorder()
	srv.ServeHTTP(recList, reqList)

	assert.Equal(t, http.StatusOK, recList.Code)

	var campaigns []db.Campaign
	err = json.Unmarshal(recList.Body.Bytes(), &campaigns)
	require.NoError(t, err)
	assert.Len(t, campaigns, 1)

	// 3. Update Campaign Status
	statusBody := map[string]string{"status": "queued"}
	statusBytes, _ := json.Marshal(statusBody)

	reqStatus := httptest.NewRequest("PATCH", "/api/v1/campaigns/1/status", bytes.NewBuffer(statusBytes))
	reqStatus.Header.Set("Content-Type", "application/json")
	recStatus := httptest.NewRecorder()
	srv.ServeHTTP(recStatus, reqStatus)

	assert.Equal(t, http.StatusOK, recStatus.Code)
}

func TestCSVUpload(t *testing.T) {
	srv, database := setupTestServer(t)

	// Seed campaign
	camp := db.Campaign{Title: "CSV Test", Subject: "Subj", BodyTemplate: "Body"}
	database.Create(&camp)

	// Build multipart CSV body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("campaign_id", "1")

	part, err := writer.CreateFormFile("file", "recipients.csv")
	require.NoError(t, err)

	csvContent := "email,name\nalice@example.com,Alice\nbob@example.com,Bob\n"
	_, _ = part.Write([]byte(csvContent))
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/campaigns/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res CSVUploadResponse
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	assert.Equal(t, 2, res.TotalParsed)
	assert.Equal(t, 2, res.TotalQueued)
	assert.Equal(t, 0, res.InvalidCount)
}

func TestDLQEndpoints(t *testing.T) {
	srv, database := setupTestServer(t)

	// Seed DLQ record
	dlq := db.DLQRecord{
		JobID:          "job_123",
		CampaignID:     1,
		RecipientEmail: "failed@example.com",
		ErrorReason:    "SMTP connection timeout",
		PayloadJSON:    `{"campaign_id":1,"recipient_id":1,"recipient_name":"Failed User","recipient_email":"failed@example.com","subject_tmpl":"Subj","body_tmpl":"Body"}`,
		Status:         db.DLQStatusPending,
	}
	database.Create(&dlq)

	// GET /api/v1/dlq
	req := httptest.NewRequest("GET", "/api/v1/dlq", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var dlqRecords []db.DLQRecord
	err := json.Unmarshal(rec.Body.Bytes(), &dlqRecords)
	require.NoError(t, err)
	assert.Len(t, dlqRecords, 1)

	// POST /api/v1/dlq/1/replay
	reqReplay := httptest.NewRequest("POST", "/api/v1/dlq/1/replay", nil)
	recReplay := httptest.NewRecorder()
	srv.ServeHTTP(recReplay, reqReplay)

	assert.Equal(t, http.StatusOK, recReplay.Code)
}
