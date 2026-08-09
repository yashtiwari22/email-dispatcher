package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/yashtiwari22/email-dispatcher/backend/db"
	"gorm.io/gorm"
)

const (
	TypeEmailDispatch = "email:dispatch"
)

// EmailTaskPayload holds details required to execute an email dispatch task.
type EmailTaskPayload struct {
	CampaignID     uint   `json:"campaign_id"`
	RecipientID    uint   `json:"recipient_id"`
	RecipientName  string `json:"recipient_name"`
	RecipientEmail string `json:"recipient_email"`
	SubjectTmpl    string `json:"subject_tmpl"`
	BodyTmpl       string `json:"body_tmpl"`
}

// NewEmailDispatchTask creates a new Asynq task for dispatching an email.
func NewEmailDispatchTask(payload EmailTaskPayload) (*asynq.Task, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}
	return asynq.NewTask(TypeEmailDispatch, bytes, asynq.MaxRetry(3), asynq.Timeout(30*time.Second)), nil
}

// WorkerProcessor handles execution of background email tasks.
type WorkerProcessor struct {
	db             *gorm.DB
	smtp           SMTPSender
	templateEngine *TemplateEngine
}

// NewWorkerProcessor initializes a new WorkerProcessor.
func NewWorkerProcessor(database *gorm.DB, smtpSender SMTPSender, tmplEngine *TemplateEngine) *WorkerProcessor {
	return &WorkerProcessor{
		db:             database,
		smtp:           smtpSender,
		templateEngine: tmplEngine,
	}
}

// ProcessTask handles an incoming Asynq task.
func (wp *WorkerProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	start := time.Now()

	var payload EmailTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	// 1. Render Subject and Body using TemplateEngine
	data := map[string]string{
		"Name":  payload.RecipientName,
		"Email": payload.RecipientEmail,
	}

	subjectKey := fmt.Sprintf("subj_%d", payload.CampaignID)
	renderedSubject, err := wp.templateEngine.Render(subjectKey, payload.SubjectTmpl, data)
	if err != nil {
		wp.recordFailure(payload, err.Error(), time.Since(start).Milliseconds())
		return err
	}

	bodyKey := fmt.Sprintf("body_%d", payload.CampaignID)
	renderedBody, err := wp.templateEngine.Render(bodyKey, payload.BodyTmpl, data)
	if err != nil {
		wp.recordFailure(payload, err.Error(), time.Since(start).Milliseconds())
		return err
	}

	// 2. Dispatch Email via SMTPSender
	if err := wp.smtp.SendEmail(payload.RecipientEmail, renderedSubject, renderedBody); err != nil {
		wp.recordFailure(payload, err.Error(), time.Since(start).Milliseconds())
		return err
	}

	// 3. Record Success in Database
	duration := time.Since(start).Milliseconds()
	now := time.Now()

	wp.db.Model(&db.Recipient{}).Where("id = ?", payload.RecipientID).Updates(map[string]interface{}{
		"status":  db.RecipientStatusSent,
		"sent_at": &now,
	})

	wp.db.Model(&db.Campaign{}).Where("id = ?", payload.CampaignID).UpdateColumn("sent_count", gorm.Expr("sent_count + 1"))

	logEntry := db.DispatchLog{
		CampaignID:     payload.CampaignID,
		RecipientID:    payload.RecipientID,
		RecipientEmail: payload.RecipientEmail,
		WorkerID:       1,
		Status:         "SUCCESS",
		DurationMS:     duration,
	}
	wp.db.Create(&logEntry)

	log.Printf("[Worker] Successfully dispatched email to %s (Campaign #%d) in %dms", payload.RecipientEmail, payload.CampaignID, duration)
	return nil
}

func (wp *WorkerProcessor) recordFailure(payload EmailTaskPayload, errMsg string, duration int64) {
	wp.db.Model(&db.Recipient{}).Where("id = ?", payload.RecipientID).Updates(map[string]interface{}{
		"status":        db.RecipientStatusFailed,
		"error_message": errMsg,
	})

	wp.db.Model(&db.Campaign{}).Where("id = ?", payload.CampaignID).UpdateColumn("failed_count", gorm.Expr("failed_count + 1"))

	logEntry := db.DispatchLog{
		CampaignID:     payload.CampaignID,
		RecipientID:    payload.RecipientID,
		RecipientEmail: payload.RecipientEmail,
		WorkerID:       1,
		Status:         "FAILED",
		ErrorMessage:   errMsg,
		DurationMS:     duration,
	}
	wp.db.Create(&logEntry)
}

// HandleDLQ saves unrecoverable failed jobs into DLQRecord table after all retries are exhausted.
func (wp *WorkerProcessor) HandleDLQ(ctx context.Context, t *asynq.Task, err error) {
	var payload EmailTaskPayload
	_ = json.Unmarshal(t.Payload(), &payload)

	jobID := fmt.Sprintf("job_%d_%d", payload.CampaignID, payload.RecipientID)
	if rw := t.ResultWriter(); rw != nil {
		jobID = rw.TaskID()
	}

	dlq := db.DLQRecord{
		JobID:          jobID,
		CampaignID:     payload.CampaignID,
		RecipientEmail: payload.RecipientEmail,
		ErrorReason:    err.Error(),
		PayloadJSON:    string(t.Payload()),
		Status:         db.DLQStatusPending,
		FailedAt:       time.Now(),
	}

	wp.db.Create(&dlq)
	log.Printf("[DLQ] Recorded failed job for recipient %s (Campaign #%d) in Dead Letter Queue", payload.RecipientEmail, payload.CampaignID)
}

