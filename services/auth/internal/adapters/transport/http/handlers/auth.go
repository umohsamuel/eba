package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/umohsamuel/eba/services/auth/internal/services"
)

type AuthHandler struct {
	services *services.Services
}

func NewAuthHandler(svc *services.Services) *AuthHandler {
	return &AuthHandler{services: svc}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	// eba here
}
func (h *AuthHandler) Login(c *gin.Context) {
	// eba here
}
