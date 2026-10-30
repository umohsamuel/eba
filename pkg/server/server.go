package server

import (
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func New(cfg Config, handler http.Handler) *http.Server {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 1 * time.Minute
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}
