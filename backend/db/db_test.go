package db

import (
	"testing"
)

func TestAutoMigrateAndSeed(t *testing.T) {
	database, err := ConnectSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite memory db: %v", err)
	}

	err = AutoMigrate(database)
	if err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}

	err = SeedTestData(database)
	if err != nil {
		t.Fatalf("failed to seed test data: %v", err)
	}

	var campaign Campaign
	err = database.Preload("Recipients").First(&campaign).Error
	if err != nil {
		t.Fatalf("failed to query seeded campaign: %v", err)
	}

	if campaign.Title != "Initial Product Onboarding" {
		t.Errorf("expected title 'Initial Product Onboarding', got '%s'", campaign.Title)
	}

	if len(campaign.Recipients) != 5 {
		t.Errorf("expected 5 recipients, got %d", len(campaign.Recipients))
	}
}
