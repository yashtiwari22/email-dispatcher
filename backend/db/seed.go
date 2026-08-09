package db

import (
	"log"

	"gorm.io/gorm"
)

// SeedTestData populates the database with initial sample templates and campaigns.
func SeedTestData(db *gorm.DB) error {
	var count int64
	db.Model(&EmailTemplate{}).Count(&count)
	if count > 0 {
		log.Println("[DB Seed] Data already exists, skipping seed")
		return nil
	}

	// 1. Seed Template
	tmpl := EmailTemplate{
		Name:    "Welcome Campaign Template",
		Subject: "Hello {{.Name}}, Welcome to Yash Campaign!",
		Body: `Hi {{.Name}},

Thank you for signing up for our product updates. We are thrilled to have you on board!

Best regards,
The Yash Campaign Team`,
		Version: 1,
	}
	if err := db.Create(&tmpl).Error; err != nil {
		return err
	}

	// 2. Seed Campaign
	campaign := Campaign{
		Title:           "Initial Product Onboarding",
		Subject:         tmpl.Subject,
		BodyTemplate:    tmpl.Body,
		Status:          CampaignStatusQueued,
		TotalRecipients: 5,
		SentCount:       0,
		FailedCount:     0,
		Recipients: []Recipient{
			{Name: "Rahul Sharma", Email: "rahul.sharma@example.com", Status: RecipientStatusPending},
			{Name: "Priya Verma", Email: "priya.verma@example.com", Status: RecipientStatusPending},
			{Name: "Amit Kumar", Email: "amit.kumar@example.com", Status: RecipientStatusPending},
			{Name: "Sneha Singh", Email: "sneha.singh@example.com", Status: RecipientStatusPending},
			{Name: "Arjun Mehta", Email: "arjun.mehta@example.com", Status: RecipientStatusPending},
		},
	}

	if err := db.Create(&campaign).Error; err != nil {
		return err
	}

	log.Println("[DB Seed] Successfully seeded sample email template and initial onboarding campaign")
	return nil
}
