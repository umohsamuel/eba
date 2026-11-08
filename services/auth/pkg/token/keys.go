package token

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"os"
	"slices"

	"github.com/golang-jwt/jwt/v5"
)

const minRSAKeyBits = 2048

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type AccessKeys struct {
	signingKey *rsa.PrivateKey
	signingKID string
	publicKeys map[string]*rsa.PublicKey
	jwks       JWKS
}

func LoadAccessKeys(privateKeyFile string, previousPublicKeyFiles []string) (*AccessKeys, error) {
	privatePEM, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read access private key: %w", err)
	}

	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("parse access private key: %w", err)
	}

	previous := make([]*rsa.PublicKey, 0, len(previousPublicKeyFiles))
	for _, file := range previousPublicKeyFiles {
		if file == "" {
			continue
		}

		publicPEM, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read previous access public key %s: %w", file, err)
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
		if err != nil {
			return nil, fmt.Errorf("parse previous access public key %s: %w", file, err)
		}

		previous = append(previous, publicKey)
	}

	return NewAccessKeys(signingKey, previous...)
}

func NewAccessKeys(signingKey *rsa.PrivateKey, previous ...*rsa.PublicKey) (*AccessKeys, error) {
	if signingKey == nil {
		return nil, errors.New("access signing key is required")
	}

	keys := &AccessKeys{
		signingKey: signingKey,
		publicKeys: make(map[string]*rsa.PublicKey, 1+len(previous)),
	}

	kid, err := keys.add(&signingKey.PublicKey)
	if err != nil {
		return nil, err
	}
	keys.signingKID = kid

	for _, publicKey := range previous {
		if _, err := keys.add(publicKey); err != nil {
			return nil, err
		}
	}

	return keys, nil
}

func (k *AccessKeys) JWKS() JWKS {
	return JWKS{Keys: slices.Clone(k.jwks.Keys)}
}

func (k *AccessKeys) add(publicKey *rsa.PublicKey) (string, error) {
	if publicKey.N.BitLen() < minRSAKeyBits {
		return "", fmt.Errorf("access key must be at least %d bits, got %d", minRSAKeyBits, publicKey.N.BitLen())
	}

	jwk := rsaJWK(publicKey)
	if _, exists := k.publicKeys[jwk.Kid]; exists {
		return jwk.Kid, nil
	}

	k.publicKeys[jwk.Kid] = publicKey
	k.jwks.Keys = append(k.jwks.Keys, jwk)

	return jwk.Kid, nil
}

func (k *AccessKeys) keyFunc(t *jwt.Token) (any, error) {
	kid, _ := t.Header["kid"].(string)

	publicKey, ok := k.publicKeys[kid]
	if !ok {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}

	return publicKey, nil
}

func rsaJWK(publicKey *rsa.PublicKey) JWK {
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())
	thumbprint := sha256.Sum256([]byte(`{"e":"` + e + `","kty":"RSA","n":"` + n + `"}`))

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: jwt.SigningMethodRS256.Alg(),
		Kid: base64.RawURLEncoding.EncodeToString(thumbprint[:]),
		N:   n,
		E:   e,
	}
}
