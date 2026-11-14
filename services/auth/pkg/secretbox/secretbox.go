package secretbox

type SecretBox interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// Implement it with AES-256-GCM in services/auth/pkg/secretbox, keyed by MFA_ENCRYPTION_KEY. Prefix each encrypted value with a key ID byte, so you can rotate the key later without breaking existing secrets.
