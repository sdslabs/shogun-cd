package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha256"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type EncryptionService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

const (
	iterations = 600000
	keyLen     = 32 // 32bytes requires for aes256
)

var (
	ErrEncryption = errors.New("Error Encrypting Data")
	ErrDecryption = errors.New("Error Decrypting Data")
)

type service struct {
	aead   cipher.AEAD
	logger utils.Logger
}

func NewService(cfg *config.Config, l utils.Logger) (EncryptionService, error) {
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

	return &service{
		aead:   aead,
		logger: l,
	}, nil
}

func deriveKey(cfg *config.Config) error {
	key, err := pbkdf2.Key(sha256.New, cfg.DBConfig.MasterKey, []byte(cfg.DBConfig.EncryptionSalt), iterations, keyLen)
	if err != nil {
		return err
	}
	cfg.DBConfig.DerivedKey = key

	return nil
}
