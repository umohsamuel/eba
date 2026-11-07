package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/umohsamuel/eba/pkg/response" // referenced by swag annotations
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	_ "github.com/umohsamuel/eba/services/auth/internal/domain/user" // referenced by swag annotations
	"github.com/umohsamuel/eba/services/auth/internal/services"
	_ "github.com/umohsamuel/eba/services/auth/internal/services/auth" // referenced by swag annotations
)

type AuthHandler struct {
	services *services.Services
	cookies  cookieWriter
}

func NewAuthHandler(svc *services.Services, cookieCfg CookieConfig) *AuthHandler {
	return &AuthHandler{services: svc, cookies: cookieWriter{cfg: cookieCfg}}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.Refresh)
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates a user account with an email and password.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerRequest							true	"Registration details"
//	@Success		201		{object}	response.Envelope[user.User]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		409		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
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

// Login godoc
//
//	@Summary		Log in
//	@Description	Authenticates a user and returns access and refresh tokens.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRequest							true	"Login credentials"
//	@Success		200		{object}	response.Envelope[auth.LoginResponse]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
		return
	}

	loginResponse, err := h.services.Auth.Login(c, req.Email, req.Password)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	h.cookies.SetAuth(c, loginResponse.Token)

	ginresponse.HandleSuccess(c, http.StatusOK, loginResponse, "success").Send()
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Exchanges a valid refresh token for a new access token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		refreshRequest							true	"Refresh token"
//	@Success		200		{object}	response.Envelope[auth.TokenConfig]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
		return
	}

	refreshResponse, err := h.services.Auth.RefreshAccessToken(c, req.RefreshToken)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	h.cookies.SetAuth(c, refreshResponse)

	ginresponse.HandleSuccess(c, http.StatusOK, refreshResponse, "success").Send()
}
