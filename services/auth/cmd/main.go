package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/umohsamuel/eba/pkg/database"
	"github.com/umohsamuel/eba/pkg/logger"
	"github.com/umohsamuel/eba/pkg/server"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/migrations"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/repository"
	httpAdapter "github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http"
	"github.com/umohsamuel/eba/services/auth/internal/services"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
)

var envCfg *env.Env

func init() {

	cfg, err := env.LoadEnv()
	if err != nil {
		log.Fatal("failed to load env")
	}

	envCfg = cfg

	log := logger.New("auth", envCfg.AppEnv, envCfg.LogLevel)
	slog.SetDefault(log)
}

func main() {

	dbCfg := database.Config{
		Host:     envCfg.DB.DB_HOST,
		Port:     envCfg.DB.DB_PORT,
		User:     envCfg.DB.DB_USER,
		Password: envCfg.DB.DB_PASSWORD,
		Database: envCfg.DB.DB_DATABASE,
		Schema:   envCfg.DB.DB_SCHEMA,
	}

	dbSrv := database.New(dbCfg)
	defer dbSrv.Close()

	db := dbSrv.DB()

	repos := repository.New(repository.Config{
		DB: db,
	})

	appServices := services.New(services.Config{
		Repository: repos,
		Env:        *envCfg,
	})

	migrator, err := migrations.New(db)
	if err != nil {
		slog.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}

	httpSrv := httpAdapter.NewHttpServer(appServices, httpAdapter.CookieConfig{
		Secure:      envCfg.Cookie.COOKIE_SECURE,
		Domain:      envCfg.Cookie.COOKIE_DOMAIN,
		RefreshPath: envCfg.Cookie.COOKIE_REFRESH_PATH,
	}, httpAdapter.MigrationConfig{
		Migrator: migrator,
		Token:    envCfg.Migration.MIGRATION_TOKEN,
	})

	srvCfg := server.Config{
		Port:         envCfg.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	appServer := server.New(srvCfg, httpSrv.Handler())

	if err := appServer.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
