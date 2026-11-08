package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

type TokenManager struct {
	issuer string

	accessKeys     *AccessKeys
	accessAudience []string

	refreshSecret   []byte
	refreshAudience []string
}

func NewTokenManager(issuer string, accessKeys *AccessKeys, accessAudience []string, refreshSecret []byte, refreshAudience []string) *TokenManager {
	return &TokenManager{
		issuer:          issuer,
		accessKeys:      accessKeys,
		accessAudience:  accessAudience,
		refreshSecret:   refreshSecret,
		refreshAudience: refreshAudience,
	}
}

func (m *TokenManager) GenerateAccessToken(user user.User, issuedAt, expiresAt time.Time) (string, error) {
	claims := CustomClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),              // sub
			ExpiresAt: jwt.NewNumericDate(expiresAt), // exp
			IssuedAt:  jwt.NewNumericDate(issuedAt),  // iat
			NotBefore: jwt.NewNumericDate(issuedAt),  // nbf
			Issuer:    m.issuer,                      // iss
			Audience:  m.accessAudience,              // aud
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.accessKeys.signingKID

	tokenString, err := token.SignedString(m.accessKeys.signingKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type CustomClaims struct {
	User user.User `json:"user"`
	jwt.RegisteredClaims
}

func (m *TokenManager) GenerateRefreshToken(user user.User, refreshTokenId string, issuedAt, expiresAt time.Time) (string, error) {

	claims := CustomClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenId,                // jti
			Subject:   user.ID.String(),              // sub
			ExpiresAt: jwt.NewNumericDate(expiresAt), // exp
			IssuedAt:  jwt.NewNumericDate(issuedAt),  // iat
			NotBefore: jwt.NewNumericDate(issuedAt),  // nbf
			Issuer:    m.issuer,                      // iss
			Audience:  m.refreshAudience,             // aud
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *TokenManager) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	return m.parse(tokenString, m.accessKeys.keyFunc, jwt.SigningMethodRS256, m.accessAudience)
}

func (m *TokenManager) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	refreshKey := func(*jwt.Token) (any, error) { return m.refreshSecret, nil }
	return m.parse(tokenString, refreshKey, jwt.SigningMethodHS256, m.refreshAudience)
}

func (m *TokenManager) AccessJWKS() JWKS {
	return m.accessKeys.JWKS()
}

func (m *TokenManager) parse(tokenString string, keyFunc jwt.Keyfunc, method jwt.SigningMethod, audience []string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims,
		keyFunc,
		jwt.WithValidMethods([]string{method.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(audience[0]),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
