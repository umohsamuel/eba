package env

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Env struct {
	AppEnv   string `env:"APP_ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	AppName  string `env:"APP_NAME" envDefault:"EBA"`

	DB    PostgresEnv
	Kafka KafkaEnv
	SMTP  SMTPEnv
}

type PostgresEnv struct {
	DB_USER     string `env:"DB_USER"`
	DB_PASSWORD string `env:"DB_PASSWORD"`
	DB_DATABASE string `env:"DB_DATABASE"`
	DB_HOST     string `env:"DB_HOST"`
	DB_SCHEMA   string `env:"DB_SCHEMA"`
	DB_PORT     int    `env:"DB_PORT" envDefault:"5432"`
	DB_SSLMODE  string `env:"DB_SSLMODE"`
}

type KafkaEnv struct {
	KAFKA_BROKERS           []string        `env:"KAFKA_BROKERS,notEmpty"`
	KAFKA_RETRY_DELAYS      []time.Duration `env:"KAFKA_RETRY_DELAYS" envDefault:"1m,5m,10m"`
	KAFKA_TOPIC_PARTITIONS  int32           `env:"KAFKA_TOPIC_PARTITIONS" envDefault:"3"`
	KAFKA_TOPIC_REPLICATION int16           `env:"KAFKA_TOPIC_REPLICATION" envDefault:"1"`
}

type SMTPEnv struct {
	SMTP_HOST       string `env:"SMTP_HOST,notEmpty"`
	SMTP_PORT       int    `env:"SMTP_PORT" envDefault:"587"`
	SMTP_USERNAME   string `env:"SMTP_USERNAME"`
	SMTP_PASSWORD   string `env:"SMTP_PASSWORD"`
	SMTP_FROM       string `env:"SMTP_FROM,notEmpty"`
	SMTP_TLS_POLICY string `env:"SMTP_TLS_POLICY" envDefault:"mandatory"`
}

func LoadEnv() (*Env, error) {
	var cfg Env

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	if len(cfg.Kafka.KAFKA_RETRY_DELAYS) != 3 {
		return nil, fmt.Errorf("KAFKA_RETRY_DELAYS must have exactly 3 values, got %d", len(cfg.Kafka.KAFKA_RETRY_DELAYS))
	}

	return &cfg, nil
}
