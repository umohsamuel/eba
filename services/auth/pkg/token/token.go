package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
)

type TokenManager struct {
	issuer string

	accessSecret   []byte
	accessAudience []string
	accessTTL      time.Duration

	refreshSecret   []byte
	refreshAudience []string
	refreshTTL      time.Duration
}

func NewTokenManager(issuer string, accessSecret []byte, accessAudience []string, accessTTL time.Duration, refreshSecret []byte, refreshAudience []string, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		issuer:          issuer,
		accessSecret:    accessSecret,
		accessAudience:  accessAudience,
		accessTTL:       accessTTL,
		refreshSecret:   refreshSecret,
		refreshAudience: refreshAudience,
		refreshTTL:      refreshTTL,
	}
}

func (m *TokenManager) GenerateAccessToken(user user.User) (string, error) {
	now := time.Now().UTC()

	claims := CustomClaims{

		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),                          // sub
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)), // exp
			IssuedAt:  jwt.NewNumericDate(now),                   // iat
			NotBefore: jwt.NewNumericDate(now),                   // nbf
			Issuer:    m.issuer,                                  // iss
			Audience:  m.refreshAudience,                         // aud
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type CustomClaims struct {
	User user.User `json:"user"`
	jwt.RegisteredClaims
}

func (m *TokenManager) GenerateRefreshToken(user user.User, refreshTokenId string) (string, error) {
	now := time.Now().UTC()

	claims := CustomClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenId,                            // jti
			Subject:   user.ID.String(),                          // sub
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)), // exp
			IssuedAt:  jwt.NewNumericDate(now),                   // iat
			NotBefore: jwt.NewNumericDate(now),                   // nbf
			Issuer:    m.issuer,                                  // iss
			Audience:  m.refreshAudience,                         // aud
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *TokenManager) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}
		return m.accessSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (m *TokenManager) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}
		return m.refreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
