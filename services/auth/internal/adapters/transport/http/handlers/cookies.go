package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umohsamuel/eba/services/auth/internal/services/auth"
)

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

type CookieConfig struct {
	Secure      bool
	Domain      string // empty = host-only cookie
	RefreshPath string
}

type cookieWriter struct {
	cfg CookieConfig
}

func (w cookieWriter) SetAuth(c *gin.Context, t *auth.TokenConfig) {
	now := time.Now()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    t.AccessToken,
		Path:     "/",
		Domain:   w.cfg.Domain,
		MaxAge:   maxAge(now, t.AccessTokenExpiresAt),
		HttpOnly: true,
		Secure:   w.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    t.RefreshToken,
		Path:     w.cfg.RefreshPath,
		Domain:   w.cfg.Domain,
		MaxAge:   maxAge(now, t.RefreshTokenExpiresAt),
		HttpOnly: true,
		Secure:   w.cfg.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (w cookieWriter) Clear(c *gin.Context) {
	for _, ck := range []struct{ name, path string }{
		{accessTokenCookie, "/"},
		{refreshTokenCookie, w.cfg.RefreshPath},
	} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     ck.name,
			Value:    "",
			Path:     ck.path,
			Domain:   w.cfg.Domain,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   w.cfg.Secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

// maxAge returns the seconds until expiresAt, or -1 (delete) if it has passed.
// A MaxAge of 0 would mean "no Max-Age attribute" in net/http, which makes it
// a session cookie, so an expired token must never produce 0.
func maxAge(now, expiresAt time.Time) int {
	secs := int(expiresAt.Sub(now).Seconds())
	if secs <= 0 {
		return -1
	}
	return secs
}
