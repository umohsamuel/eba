package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) JWKS(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, h.services.Auth.JWKS())
}
