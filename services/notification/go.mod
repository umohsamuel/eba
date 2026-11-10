module github.com/umohsamuel/eba/services/notification

go 1.27.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.12.0 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.31 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

require (
	github.com/IBM/sarama v1.61.1
	github.com/caarlos0/env/v11 v11.4.2
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/umohsamuel/eba/pkg/database v0.0.0
	github.com/umohsamuel/eba/pkg/events v0.0.0
	github.com/umohsamuel/eba/pkg/logger v0.0.0
	github.com/wneessen/go-mail v0.8.2
)

replace (
	github.com/umohsamuel/eba/pkg/database => ../../pkg/database
	github.com/umohsamuel/eba/pkg/events => ../../pkg/events
	github.com/umohsamuel/eba/pkg/logger => ../../pkg/logger
)
