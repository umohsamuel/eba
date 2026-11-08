package middlewares

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/pkg/token"
)

const (
	AccessTokenCookie = "access_token"
	claimsKey         = "auth.claims"
)

type AccessTokenVerifier interface {
	VerifyAccessToken(ctx context.Context, accessToken string) (*token.CustomClaims, error)
}

type Auth struct {
	verifier AccessTokenVerifier
}

func NewAuth(verifier AccessTokenVerifier) *Auth {
	return &Auth{verifier: verifier}
}

func (m *Auth) Handle(c *gin.Context) {
	accessToken, err := accessTokenFrom(c)
	if err != nil {
		ginresponse.HandleError(c, err).Abort()
		return
	}

	claims, err := m.verifier.VerifyAccessToken(c, accessToken)
	if err != nil {
		ginresponse.HandleError(c, err).Abort()
		return
	}

	c.Set(claimsKey, claims)
	c.Next()
}

func ClaimsFrom(c *gin.Context) (*token.CustomClaims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*token.CustomClaims)
	return claims, ok
}

func accessTokenFrom(c *gin.Context) (string, error) {
	if header := c.GetHeader("Authorization"); header != "" {
		scheme, accessToken, ok := strings.Cut(header, " ")
		accessToken = strings.TrimSpace(accessToken)
		if !ok || !strings.EqualFold(scheme, "Bearer") || accessToken == "" {
			return "", domainerrors.ErrTokenInvalid
		}
		return accessToken, nil
	}

	if accessToken, err := c.Cookie(AccessTokenCookie); err == nil && accessToken != "" {
		return accessToken, nil
	}

	return "", domainerrors.ErrTokenInvalid
}
