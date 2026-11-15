package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const keySize = 32

var ErrInvalidCiphertext = errors.New("secretbox: invalid ciphertext")

type SecretBox interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type AESGCM struct {
	keyID byte
	aead  cipher.AEAD
}

var _ SecretBox = (*AESGCM)(nil)

func NewAESGCM(keyID byte, key []byte) (*AESGCM, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("secretbox: key must be %d bytes, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secretbox: create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secretbox: create gcm: %w", err)
	}

	return &AESGCM{keyID: keyID, aead: aead}, nil
}

func NewAESGCMFromBase64(keyID byte, encodedKey string) (*AESGCM, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("secretbox: decode key: %w", err)
	}

	return NewAESGCM(keyID, key)
}

func (b *AESGCM) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("secretbox: generate nonce: %w", err)
	}

	out := make([]byte, 0, 1+len(nonce)+len(plaintext)+b.aead.Overhead())
	out = append(out, b.keyID)
	out = append(out, nonce...)

	return b.aead.Seal(out, nonce, plaintext, []byte{b.keyID}), nil
}

func (b *AESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := b.aead.NonceSize()
	if len(ciphertext) < 1+nonceSize+b.aead.Overhead() || ciphertext[0] != b.keyID {
		return nil, ErrInvalidCiphertext
	}

	nonce := ciphertext[1 : 1+nonceSize]
	sealed := ciphertext[1+nonceSize:]

	plaintext, err := b.aead.Open(nil, nonce, sealed, ciphertext[:1])
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}
