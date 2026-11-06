package main

import (
	"net/http"

	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/umohsamuel/eba/pkg/server"
	"github.com/umohsamuel/eba/pkg/shutdown"
	env "github.com/umohsamuel/eba/services/api/pkg"
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

	engine := gin.Default()

	r := RegisterRoutes(engine)

	srvCfg := server.Config{
		Port: envCfg.Port,
	}

	server := server.New(srvCfg, r)

	done := make(chan bool, 1)
	go shutdown.Shutdown(server, done)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}

func RegisterRoutes(r *gin.Engine) http.Handler {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{envCfg.CLIENT_ORIGIN},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	return r
}
