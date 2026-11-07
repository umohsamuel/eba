package env

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Env struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	CLIENT_ORIGIN string `env:"CLIENT_ORIGIN"`
	AppEnv        string `env:"APP_ENV" envDefault:"development"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	DB            PostgresEnv
	JWT           JWTEnv
	Cookie        CookieEnv
	Migration     MigrationEnv
}

type MigrationEnv struct {
	MIGRATION_TOKEN string `env:"MIGRATION_TOKEN"`
}

type JWTEnv struct {
	JWT_SECRET         string   `env:"JWT_SECRET"`
	JWT_REFRESH_SECRET string   `env:"JWT_REFRESH_SECRET"`
	JWT_ISSUER         string   `env:"JWT_ISSUER"`
	JWT_AUDIENCE       []string `env:"JWT_AUDIENCE"`

	JWT_ACCESS_TTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWT_REFRESH_TTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
}

type CookieEnv struct {
	COOKIE_SECURE       bool   `env:"COOKIE_SECURE" envDefault:"true"`
	COOKIE_DOMAIN       string `env:"COOKIE_DOMAIN"` // empty = host-only
	COOKIE_REFRESH_PATH string `env:"COOKIE_REFRESH_PATH" envDefault:"/api/v1/auth"`
}

type PostgresEnv struct {
	DB_USER     string `env:"DB_USER"`
	DB_PASSWORD string `env:"DB_PASSWORD"`
	DB_DATABASE string `env:"DB_DATABASE"`
	DB_HOST     string `env:"DB_HOST"`
	DB_SCHEMA   string `env:"DB_SCHEMA"`
	DB_PORT     int    `env:"DB_PORT" envDefault:"5432"`
}

func LoadEnv() (*Env, error) {
	var cfg Env

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	return &cfg, nil
}
