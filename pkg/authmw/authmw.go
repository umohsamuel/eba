package authmw

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	apperror "github.com/umohsamuel/eba/pkg/errors"
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
)

const claimsKey = "authmw.claims"

var (
	ErrTokenInvalid = apperror.New(apperror.KindUnauthorized, "invalid token")
	ErrTokenExpired = apperror.New(apperror.KindUnauthorized, "token expired")
)

type Config struct {
	PublicKey  *rsa.PublicKey
	Issuer     string
	Audience   string
	CookieName string
}

type Claims struct {
	jwt.RegisteredClaims
}

func (c *Claims) UserID() string {
	return c.Subject
}

type Middleware struct {
	cfg    Config
	parser *jwt.Parser
}

func New(cfg Config) (*Middleware, error) {
	if cfg.PublicKey == nil {
		return nil, errors.New("authmw: public key is required")
	}
	if cfg.Issuer == "" {
		return nil, errors.New("authmw: issuer is required")
	}
	if cfg.Audience == "" {
		return nil, errors.New("authmw: audience is required")
	}

	return &Middleware{
		cfg: cfg,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
			jwt.WithIssuer(cfg.Issuer),
			jwt.WithAudience(cfg.Audience),
			jwt.WithExpirationRequired(),
		),
	}, nil
}

func LoadPublicKey(file string) (*rsa.PublicKey, error) {
	pem, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("authmw: read public key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("authmw: parse public key: %w", err)
	}

	return publicKey, nil
}

func (m *Middleware) Handle(c *gin.Context) {
	tokenString, err := m.tokenFrom(c)
	if err != nil {
		ginresponse.HandleError(c, err).Abort()
		return
	}

	claims := &Claims{}
	_, err = m.parser.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		return m.cfg.PublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			ginresponse.HandleError(c, ErrTokenExpired.Wrap(err)).Abort()
			return
		}
		ginresponse.HandleError(c, ErrTokenInvalid.Wrap(err)).Abort()
		return
	}

	c.Set(claimsKey, claims)
	c.Next()
}

func ClaimsFrom(c *gin.Context) (*Claims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*Claims)
	return claims, ok
}

func (m *Middleware) tokenFrom(c *gin.Context) (string, error) {
	if header := c.GetHeader("Authorization"); header != "" {
		scheme, token, ok := strings.Cut(header, " ")
		token = strings.TrimSpace(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			return "", ErrTokenInvalid
		}
		return token, nil
	}

	if m.cfg.CookieName != "" {
		if token, err := c.Cookie(m.cfg.CookieName); err == nil && token != "" {
			return token, nil
		}
	}

	return "", ErrTokenInvalid
}
