package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http/handlers"
	"github.com/umohsamuel/eba/services/auth/internal/services"
)

type HttpServer struct {
	router   *gin.Engine
	services *services.Services
}

func NewHttpServer(svc *services.Services) *HttpServer {
	r := gin.Default()
	s := &HttpServer{router: r, services: svc}

	s.setupRoutes()
	return s
}

func (s *HttpServer) setupRoutes() {
	v1 := s.router.Group("/api/v1")

	authHandler := handlers.NewAuthHandler(s.services)
	authHandler.RegisterRoutes(v1)
}

func (s *HttpServer) Handler() http.Handler {
	return s.router
}
