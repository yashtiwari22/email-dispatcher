package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/yashtiwari22/email-dispatcher/backend/db"
)

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "email_dispatcher")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	cfg := db.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	}

	database, err := db.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to run auto-migrations: %v", err)
	}

	// Seed test data if database is empty
	if err := db.SeedTestData(database); err != nil {
		log.Printf("Failed to seed initial test data: %v", err)
	}

	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	server := NewServer(database, asynqClient)

	port := getEnvOrDefault("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("[API Gateway] Server starting on HTTP %s...", addr)

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("API Server failed: %v", err)
	}
}
