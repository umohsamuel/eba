package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperror "github.com/umohsamuel/eba/pkg/errors"
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
		authGroup.POST("/forgot-password", h.ForgotPassword)
		authGroup.POST("/reset-password", h.ResetPassword)
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
//	@Description	Authenticates a user. Web clients (default) receive the tokens as HttpOnly cookies and only
//	@Description	their expiry times in the body. Clients sending X-Client-Type: mobile receive the tokens in
//	@Description	the body and no cookies.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-Type	header		string									false	"Client type"	Enums(web, mobile)
//	@Param			request			body		loginRequest							true	"Login credentials"
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

	if clientTypeOf(c) == clientWeb {
		h.cookies.SetAuth(c, loginResponse.Token)
		loginResponse.Token = withoutTokens(loginResponse.Token)
	}

	ginresponse.HandleSuccess(c, http.StatusOK, loginResponse, "success").Send()
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Exchanges a valid refresh token for a new access token. Web clients (default) send no body;
//	@Description	the refresh token is read from its cookie and the new access token is set as a cookie.
//	@Description	Clients sending X-Client-Type: mobile send the refresh token in the body and receive the
//	@Description	tokens in the response.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-Type	header		string									false	"Client type"	Enums(web, mobile)
//	@Param			request			body		refreshRequest							false	"Refresh token (mobile clients only)"
//	@Success		200		{object}	response.Envelope[auth.TokenConfig]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	client := clientTypeOf(c)

	// Each client type reads the token from exactly one place. Mobile must
	// never fall back to the cookie, or a browser script could send
	// X-Client-Type: mobile and receive the cookie's tokens in the body.
	var rawRefreshToken string
	if client == clientMobile {
		var req refreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
			return
		}
		rawRefreshToken = req.RefreshToken
	} else {
		rawRefreshToken, _ = c.Cookie(refreshTokenCookie) // a missing cookie leaves it empty
	}

	if rawRefreshToken == "" {
		ginresponse.HandleError(c, domainerrors.ErrTokenInvalid).Send()
		return
	}

	refreshResponse, err := h.services.Auth.RefreshAccessToken(c, rawRefreshToken)
	if err != nil {
		// Drop cookies holding a token that can no longer be used, so the
		// browser stops sending it.
		if client == clientWeb && apperror.ErrorKind(err) == apperror.KindUnauthorized {
			h.cookies.Clear(c)
		}
		ginresponse.HandleError(c, err).Send()
		return
	}

	if client == clientWeb {
		h.cookies.SetAuth(c, refreshResponse)
		refreshResponse = withoutTokens(refreshResponse)
	}

	ginresponse.HandleSuccess(c, http.StatusOK, refreshResponse, "success").Send()
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

const forgotPasswordMessage = "If an account exists for that email, we've sent a reset link."

// ForgotPassword godoc
//
//	@Summary		Request a password reset
//	@Description	Sends a password reset link if an account exists for the email. The response is the same
//	@Description	whether or not the account exists.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		forgotPasswordRequest	true	"Account email"
//	@Success		200		{object}	response.Envelope[any]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
		return
	}

	if err := h.services.Auth.ForgotPassword(c, req.Email); err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleMessage(c, http.StatusOK, forgotPasswordMessage).Send()
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword godoc
//
//	@Summary		Reset password
//	@Description	Sets a new password using the token from the reset link. The token works once, and all
//	@Description	existing sessions are signed out.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		resetPasswordRequest	true	"Reset token and new password"
//	@Success		200		{object}	response.Envelope[any]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
		return
	}

	if err := h.services.Auth.ResetPassword(c, req.Token, req.NewPassword); err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleMessage(c, http.StatusOK, "password has been reset").Send()
}
