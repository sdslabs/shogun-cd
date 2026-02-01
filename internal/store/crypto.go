package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/kunalvirwal/shogun-cd/internal/config"
)

const (
	iterations = 100000
	keyLen     = 32 // 32bytes requires for aes256
)

type Crypto interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type crypto struct {
	aead cipher.AEAD
}

func deriveKey(cfg *config.Config) error {
	key, err := pbkdf2.Key(sha256.New, cfg.DBConfig.MasterKey, []byte(cfg.DBConfig.EncryptionSalt), iterations, keyLen)
	if err != nil {
		return err
	}
	cfg.DBConfig.DerivedKey = key

	return nil
}

func newCrypto(cfg *config.Config) (Crypto, error) {
	if err := deriveKey(cfg); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(cfg.DBConfig.DerivedKey)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &crypto{
		aead: aead,
	}, nil
}

func (c *crypto) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())

	// 12 random bytes from the OS's cryptographically secure random source
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)

	return hex.EncodeToString(ciphertext), nil
}

func (c *crypto) Decrypt(ciphertext string) (string, error) {
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
		return "", fmt.Errorf("decryption failed : %w", err)
	}

	return string(plaintext), nil
}
