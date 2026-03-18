package encryption

import (
	"encoding/hex"
	"fmt"
)

func (c *service) Decrypt(ciphertext string) (string, error) {
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %w", err)
	}

	// The data MUST contain at least the nonce, if it's smaller than std nonce size then its not a valid encrypted data
	nonceSize := c.aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	plaintext, err := c.aead.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("%w : %w", ErrDecryption, err)
	}

	return string(plaintext), nil
}
