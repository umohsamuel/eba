package main

import (
	"log"
	"time"

	"github.com/umohsamuel/eba/pkg/database"
	"github.com/umohsamuel/eba/pkg/server"
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
		Users: repos.Users,
	})

	httpSrv := httpAdapter.NewHttpServer(appServices)

	srvCfg := server.Config{
		Port:         8080,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	appServer := server.New(srvCfg, httpSrv.Handler())

	if err := appServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
