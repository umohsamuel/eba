package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/umohsamuel/eba/services/auth/internal/services/auth"
)

// clientTypeHeader lets non-browser clients opt into receiving tokens in the
// response body. Browsers never send it, so they get cookies by default.
const clientTypeHeader = "X-Client-Type"

type clientType int

const (
	// clientWeb receives tokens only as HttpOnly cookies and refreshes with
	// the refresh cookie. It is the default.
	clientWeb clientType = iota
	// clientMobile receives tokens only in the response body and refreshes
	// with the token in the request body. It never reads or sets cookies, so
	// a script in a browser cannot use this mode to read the cookie tokens.
	clientMobile
)

func clientTypeOf(c *gin.Context) clientType {
	if strings.EqualFold(c.GetHeader(clientTypeHeader), "mobile") {
		return clientMobile
	}
	return clientWeb
}

// withoutTokens returns a copy of t with the token strings removed, keeping
// the expiry times so web clients know when to refresh.
func withoutTokens(t *auth.TokenConfig) *auth.TokenConfig {
	stripped := *t
	stripped.AccessToken = ""
	stripped.RefreshToken = ""
	return &stripped
}
