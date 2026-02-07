package encryption

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func (c *service) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())

	// 12 random bytes from the OS's cryptographically secure random source
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)

	return hex.EncodeToString(ciphertext), nil
}
