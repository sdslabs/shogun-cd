package git

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"golang.org/x/crypto/ssh"
)

const (
	deployKeyComment = "shogun@sdslabs.co"
)

// CreateAndAddDeployKey reuses an existing deploy key pair or creates one for the git service.
func (s *Service) CreateAndAddDeployKey(keyDir string) error {
	privateKeyPath := filepath.Join(keyDir, "deploy_id_ed25519")
	publicKeyPath := filepath.Join(keyDir, "deploy_id_ed25519.pub")

	keys := &KeyPair{
		PrivatePath: privateKeyPath,
		PublicPath:  publicKeyPath,
	}

	privateInfo, privateErr := os.Stat(privateKeyPath)
	_, publicErr := os.Stat(publicKeyPath)
	if privateErr != nil && !errors.Is(privateErr, os.ErrNotExist) {
		return s.logAndReturnError("check private deploy key %s: %w", privateKeyPath, privateErr)
	}
	if publicErr != nil && !errors.Is(publicErr, os.ErrNotExist) {
		return s.logAndReturnError("check public deploy key %s: %w", publicKeyPath, publicErr)
	}
	if (privateErr == nil) != (publicErr == nil) {
		return s.logAndReturnError("incomplete deploy key pair at %s: restore the missing file before starting Shogun", keyDir)
	}
	if privateErr == nil {
		if !privateInfo.Mode().IsRegular() {
			return s.logAndReturnError("private deploy key %s is not a regular file", privateKeyPath)
		}
		// OpenSSH rejects private keys accessible to the group or others.
		if privateInfo.Mode().Perm()&0077 != 0 {
			if err := os.Chmod(privateKeyPath, 0600); err != nil {
				return s.logAndReturnError("secure private deploy key %s: %w (run Shogun as the key owner or fix ownership and permissions)", privateKeyPath, err)
			}
		}
		keyFile, err := os.Open(privateKeyPath)
		if err != nil {
			return s.logAndReturnError("read private deploy key %s: %w (run Shogun as the key owner or fix ownership)", privateKeyPath, err)
		}
		if err := keyFile.Close(); err != nil {
			return s.logAndReturnError("close private deploy key %s: %w", privateKeyPath, err)
		}
		pubBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return s.logAndReturnError("read public deploy key %s: %w", publicKeyPath, err)
		}
		s.logger.LogInfo("Deploy key pair already exists at %s and %s", privateKeyPath, publicKeyPath)
		s.repo.DeployKeys = keys
		s.logDeployKey(pubBytes)
		return nil
	}

	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return s.logAndReturnError("create deploy key directory %s: %w", keyDir, err)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return s.logAndReturnError("generate deploy key pair: %w", err)
	}

	// ------ PRIVATE KEY (PEM format) ------
	privKey, err := ssh.MarshalPrivateKey(priv, deployKeyComment)
	if err != nil {
		return s.logAndReturnError("marshal private deploy key: %w", err)
	}

	privPem := pem.EncodeToMemory(privKey)
	if err := os.WriteFile(privateKeyPath, privPem, 0600); err != nil {
		return s.logAndReturnError("write private deploy key %s: %w", privateKeyPath, err)
	}

	// ------ PUBLIC KEY (Authorised Keys format format) ------
	sshPub, err := ssh.NewPublicKey(pub)

	if err != nil {
		return s.logAndReturnError("generate SSH public key: %w", err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPub)
	pubBytes = bytes.Replace(pubBytes, []byte("\n"), []byte(" "+deployKeyComment+"\n"), 1)
	if err := os.WriteFile(publicKeyPath, pubBytes, 0644); err != nil {
		return s.logAndReturnError("write public deploy key %s: %w", publicKeyPath, err)
	}

	s.repo.DeployKeys = keys
	s.logger.LogInfo("Deploy key pair created at %s and %s", privateKeyPath, publicKeyPath)
	s.logDeployKey(pubBytes)
	return nil
}

func (s *Service) logAndReturnError(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	s.logger.LogNewError("%v", err)
	return err
}

func (s *Service) logDeployKey(publicKey []byte) {
	s.logger.LogInfo("Deploy Key to add to Github:")
	s.logger.LogCustom(utils.Magenta, "Deploy-Key", string(publicKey))
}
