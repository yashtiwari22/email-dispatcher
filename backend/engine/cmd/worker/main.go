package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/yashtiwari22/email-dispatcher/backend/db"
	"github.com/yashtiwari22/email-dispatcher/backend/engine"
)

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	dbHost := getEnvOrDefault("DB_HOST", "127.0.0.1")
	dbPort := getEnvOrDefault("DB_PORT", "5434")
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

	smtpHost := getEnvOrDefault("SMTP_HOST", "localhost")
	smtpPortStr := getEnvOrDefault("SMTP_PORT", "1025")
	smtpPort, _ := strconv.Atoi(smtpPortStr)
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := getEnvOrDefault("FROM_EMAIL", "noreply@dispatcher.com")

	smtpClient := engine.NewSMTPSender(smtpHost, smtpPort, smtpUser, smtpPassword, fromEmail)
	tmplEngine := engine.NewTemplateEngine()
	processor := engine.NewWorkerProcessor(database, smtpClient, tmplEngine)

	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(engine.TypeEmailDispatch, processor.ProcessTask)

	log.Printf("[Worker Engine] Asynq worker server starting listening on Redis %s...", redisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Worker Engine failed: %v", err)
	}
}
