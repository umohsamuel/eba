package main

import (
	"log"

	"github.com/umohsamuel/eba/pkg/database"
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

	srvsCfg := services.Config{
		// in order to proceed, create user repository in sqlc repo
		// Users: sqlc.,
	}
	srvs := services.New(srvsCfg)

	// q := sqlc.New(db)
}
