package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database connection parameters.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectPostgres connects to PostgreSQL using GORM and configures connection pooling.
func ConnectPostgres(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Configure connection pool settings for high performance
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("[DB] Successfully connected to PostgreSQL database")
	return db, nil
}

// ConnectSQLite creates an in-memory or file-based SQLite instance (useful for unit testing).
func ConnectSQLite(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = ":memory:"
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}
	return db, nil
}

// AutoMigrate runs schema auto-migrations for all domain entities.
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&Campaign{},
		&Recipient{},
		&EmailTemplate{},
		&DispatchLog{},
		&DLQRecord{},
	)
	if err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}
	log.Println("[DB] Database auto-migration completed successfully")
	return nil
}
