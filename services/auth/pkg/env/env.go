package env

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Env struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	CLIENT_ORIGIN string `env:"CLIENT_ORIGIN"`
	DB            PostgresEnv
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
