package db

import (
	"time"
)

// CampaignStatus constants
const (
	CampaignStatusDraft      = "draft"
	CampaignStatusQueued     = "queued"
	CampaignStatusProcessing = "processing"
	CampaignStatusCompleted  = "completed"
	CampaignStatusFailed     = "failed"
	CampaignStatusPaused     = "paused"
)

// RecipientStatus constants
const (
	RecipientStatusPending = "pending"
	RecipientStatusSent    = "sent"
	RecipientStatusFailed  = "failed"
)

// DLQStatus constants
const (
	DLQStatusPending   = "pending"
	DLQStatusReplayed  = "replayed"
	DLQStatusDiscarded = "discarded"
)

// Campaign represents an email dispatch campaign.
type Campaign struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	Title           string        `gorm:"type:varchar(255);not null" json:"title"`
	Subject         string        `gorm:"type:varchar(255);not null" json:"subject"`
	BodyTemplate    string        `gorm:"type:text;not null" json:"body_template"`
	Status          string        `gorm:"type:varchar(50);default:'draft';index" json:"status"`
	TotalRecipients int           `gorm:"default:0" json:"total_recipients"`
	SentCount       int           `gorm:"default:0" json:"sent_count"`
	FailedCount     int           `gorm:"default:0" json:"failed_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Recipients      []Recipient   `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE" json:"recipients,omitempty"`
	DispatchLogs    []DispatchLog `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE" json:"dispatch_logs,omitempty"`
}

// Recipient represents a targeted recipient in a campaign.
type Recipient struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	CampaignID   uint       `gorm:"index;not null" json:"campaign_id"`
	Name         string     `gorm:"type:varchar(255)" json:"name"`
	Email        string     `gorm:"type:varchar(255);not null;index" json:"email"`
	Status       string     `gorm:"type:varchar(50);default:'pending';index" json:"status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// EmailTemplate represents a reusable email template.
type EmailTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	Subject   string    `gorm:"type:varchar(255);not null" json:"subject"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Version   int       `gorm:"default:1" json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DispatchLog records execution results per email delivery attempt.
type DispatchLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CampaignID     uint      `gorm:"index;not null" json:"campaign_id"`
	RecipientID    uint      `gorm:"index;not null" json:"recipient_id"`
	RecipientEmail string    `gorm:"type:varchar(255);not null" json:"recipient_email"`
	WorkerID       int       `gorm:"not null" json:"worker_id"`
	Status         string    `gorm:"type:varchar(50);not null" json:"status"`
	ErrorMessage   string    `gorm:"type:text" json:"error_message,omitempty"`
	DurationMS     int64     `json:"duration_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

// DLQRecord stores failed jobs for inspection and retry/replay operations.
type DLQRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	JobID          string    `gorm:"type:varchar(255);index" json:"job_id"`
	CampaignID     uint      `gorm:"index;not null" json:"campaign_id"`
	RecipientEmail string    `gorm:"type:varchar(255);not null" json:"recipient_email"`
	ErrorReason    string    `gorm:"type:text;not null" json:"error_reason"`
	PayloadJSON    string    `gorm:"type:text;not null" json:"payload_json"`
	Status         string    `gorm:"type:varchar(50);default:'pending';index" json:"status"`
	FailedAt       time.Time `json:"failed_at"`
}
