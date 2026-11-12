package env

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Env struct {
	Port          int               `env:"PORT" envDefault:"8080"`
	CLIENT_ORIGIN string            `env:"CLIENT_ORIGIN"`
	DOCS_SPECS    map[string]string `env:"DOCS_SPECS" envKeyValSeparator:"="`
}

func LoadEnv() (*Env, error) {
	var cfg Env

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	return &cfg, nil
}
