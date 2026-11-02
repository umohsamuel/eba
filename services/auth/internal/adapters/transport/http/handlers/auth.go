package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
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
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	createdUser, err := h.services.Auth.Register(c, req.Email, req.Password)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusCreated, createdUser, "success").Send()
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	loginResponse, err := h.services.Auth.Login(c, req.Email, req.Password)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, loginResponse, "success").Send()
}
