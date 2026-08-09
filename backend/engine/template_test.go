package engine

import (
	"context"
	"testing"

	"github.com/yashtiwari22/email-dispatcher/backend/db"
)

func TestTemplateEngineAndWorker(t *testing.T) {
	// 1. Initialize in-memory SQLite DB
	database, err := db.ConnectSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to connect sqlite: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed auto migrate: %v", err)
	}

	// Seed campaign and recipient
	campaign := db.Campaign{Title: "Test Campaign", Subject: "Hello {{.Name}}", BodyTemplate: "Welcome {{.Name}}!"}
	database.Create(&campaign)

	recipient := db.Recipient{CampaignID: campaign.ID, Name: "Test User", Email: "test@example.com", Status: db.RecipientStatusPending}
	database.Create(&recipient)

	// 2. Initialize Engine Services
	tmplEngine := NewTemplateEngine()
	mockSMTP := &MockSMTPSender{}
	processor := NewWorkerProcessor(database, mockSMTP, tmplEngine)

	// 3. Create & Process Task
	payload := EmailTaskPayload{
		CampaignID:     campaign.ID,
		RecipientID:    recipient.ID,
		RecipientName:  recipient.Name,
		RecipientEmail: recipient.Email,
		SubjectTmpl:    campaign.Subject,
		BodyTmpl:       campaign.BodyTemplate,
	}

	task, err := NewEmailDispatchTask(payload)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if err := processor.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("failed to process task: %v", err)
	}

	// 4. Assertions
	if len(mockSMTP.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(mockSMTP.SentEmails))
	}

	sent := mockSMTP.SentEmails[0]
	if sent.To != "test@example.com" {
		t.Errorf("expected to 'test@example.com', got '%s'", sent.To)
	}
	if sent.Subject != "Hello Test User" {
		t.Errorf("expected subject 'Hello Test User', got '%s'", sent.Subject)
	}
	if sent.Body != "Welcome Test User!" {
		t.Errorf("expected body 'Welcome Test User!', got '%s'", sent.Body)
	}

	// Verify DB status updated
	var updatedRecipient db.Recipient
	database.First(&updatedRecipient, recipient.ID)
	if updatedRecipient.Status != db.RecipientStatusSent {
		t.Errorf("expected recipient status 'sent', got '%s'", updatedRecipient.Status)
	}
}
