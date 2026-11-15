package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/umohsamuel/eba/pkg/response"
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http/middlewares"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/services"
	_ "github.com/umohsamuel/eba/services/auth/internal/services/mfa"
)

type MFAHandler struct {
	services *services.Services
}

func NewMFAHandler(svc *services.Services) *MFAHandler {
	return &MFAHandler{services: svc}
}

func (h *MFAHandler) RegisterRoutes(rg *gin.RouterGroup) {
	mfaGroup := rg.Group("/auth/mfa")
	{
		mfaGroup.POST("/totp/setup", h.SetupTOTP)
		mfaGroup.POST("/totp/confirm", h.ConfirmTOTP)
		mfaGroup.POST("/totp/disable", h.DisableTOTP)
		mfaGroup.POST("/recovery-codes", h.RegenerateRecoveryCodes)
	}
}

type mfaCodeRequest struct {
	Code string `json:"code"`
}

type recoveryCodesResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// SetupTOTP godoc
//
//	@Summary		Start TOTP setup
//	@Description	Generates a TOTP secret for the signed-in user. Render otpauth_url as a QR code, or let the
//	@Description	user enter the secret manually. Two-factor authentication stays off until it is confirmed.
//	@Tags			mfa
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope[mfa.SetupResult]
//	@Failure		401	{object}	response.Envelope[any]
//	@Failure		409	{object}	response.Envelope[any]
//	@Failure		500	{object}	response.Envelope[any]
//	@Router			/auth/mfa/totp/setup [post]
func (h *MFAHandler) SetupTOTP(c *gin.Context) {
	userID, err := userIDFrom(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	result, err := h.services.MFA.Setup(c, userID)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, result, "success").Send()
}

// ConfirmTOTP godoc
//
//	@Summary		Confirm TOTP setup
//	@Description	Turns on two-factor authentication with a code from the authenticator app and returns
//	@Description	recovery codes. The recovery codes are shown only once.
//	@Tags			mfa
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		mfaCodeRequest	true	"Code from the authenticator app"
//	@Success		200		{object}	response.Envelope[recoveryCodesResponse]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		404		{object}	response.Envelope[any]
//	@Failure		409		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/mfa/totp/confirm [post]
func (h *MFAHandler) ConfirmTOTP(c *gin.Context) {
	userID, req, ok := userAndCode(c)
	if !ok {
		return
	}

	codes, err := h.services.MFA.Confirm(c, userID, req.Code)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, recoveryCodesResponse{RecoveryCodes: codes}, "success").Send()
}

// DisableTOTP godoc
//
//	@Summary		Turn off two-factor authentication
//	@Description	Requires a current authenticator code or a recovery code. Deletes the TOTP secret and all
//	@Description	recovery codes, and emails the user an alert.
//	@Tags			mfa
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		mfaCodeRequest	true	"Authenticator or recovery code"
//	@Success		200		{object}	response.Envelope[any]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		404		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/mfa/totp/disable [post]
func (h *MFAHandler) DisableTOTP(c *gin.Context) {
	userID, req, ok := userAndCode(c)
	if !ok {
		return
	}

	if err := h.services.MFA.Disable(c, userID, req.Code); err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleMessage(c, http.StatusOK, "two-factor authentication disabled").Send()
}

// RegenerateRecoveryCodes godoc
//
//	@Summary		Regenerate recovery codes
//	@Description	Requires a current authenticator code or a recovery code. Replaces all recovery codes; the new
//	@Description	codes are shown only once.
//	@Tags			mfa
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		mfaCodeRequest	true	"Authenticator or recovery code"
//	@Success		200		{object}	response.Envelope[recoveryCodesResponse]
//	@Failure		400		{object}	response.Envelope[any]
//	@Failure		401		{object}	response.Envelope[any]
//	@Failure		404		{object}	response.Envelope[any]
//	@Failure		500		{object}	response.Envelope[any]
//	@Router			/auth/mfa/recovery-codes [post]
func (h *MFAHandler) RegenerateRecoveryCodes(c *gin.Context) {
	userID, req, ok := userAndCode(c)
	if !ok {
		return
	}

	codes, err := h.services.MFA.RegenerateRecoveryCodes(c, userID, req.Code)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, recoveryCodesResponse{RecoveryCodes: codes}, "success").Send()
}

func userIDFrom(c *gin.Context) (uuid.UUID, error) {
	claims, ok := middlewares.ClaimsFrom(c)
	if !ok {
		return uuid.Nil, domainerrors.ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, domainerrors.ErrTokenInvalid.Wrap(err)
	}

	return userID, nil
}

func userAndCode(c *gin.Context) (uuid.UUID, mfaCodeRequest, bool) {
	var req mfaCodeRequest

	userID, err := userIDFrom(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return uuid.Nil, req, false
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginresponse.HandleError(c, domainerrors.ErrBadRequest.Wrap(err)).Send()
		return uuid.Nil, req, false
	}

	return userID, req, true
}
