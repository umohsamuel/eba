package adapters

import (
	"github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
)

type Adapters struct {
	http http.HttpServer
}

type Config struct {
	Env env.Env
}

func New()
