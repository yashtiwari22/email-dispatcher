module github.com/yashtiwari22/email-dispatcher/backend/api

go 1.26.5

require (
	github.com/hibiken/asynq v0.26.0
	github.com/yashtiwari22/email-dispatcher/backend/db v0.0.0-00010101000000-000000000000
	gorm.io/gorm v1.31.2
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/redis/go-redis/v9 v9.14.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gorm.io/driver/postgres v1.6.2 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
)

replace (
	github.com/yashtiwari22/email-dispatcher/backend/db => ../db
	github.com/yashtiwari22/email-dispatcher/backend/engine => ../engine
)
