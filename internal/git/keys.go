package git

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"golang.org/x/crypto/ssh"
)

const (
	defaultKeyDir    = ".data/.ssh"
	deployKeyComment = "shogun@sdslabs.co"
)

// CreateAndAddDeployKey creates an SSH key pair to be added as a deploy key and and adds it to git service
func (s *Service) CreateAndAddDeployKey(keyDir string) error {

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		s.logger.LogNewError("Failed to generate ed25519 key pair: %v", err)
		return err
	}

	if keyDir == "" {
		keyDir = defaultKeyDir
	}

	keyDir = strings.TrimSuffix(keyDir, "/")

	// Key file names
	privateKeyPath := keyDir + "/deploy_id_ed25519"
	publicKeyPath := keyDir + "/deploy_id_ed25519.pub"

	keys := &KeyPair{
		PrivatePath: privateKeyPath,
		PublicPath:  publicKeyPath,
	}

	// Log the public key for user to add to Github
	defer func() {
		if f, err := os.ReadFile(publicKeyPath); err == nil {
			s.logger.LogInfo("Deploy Key to add to Github:")
			s.logger.LogCustom(utils.Magenta, "Deploy-Key", string(f))
			return
		}
		s.logger.LogNewError("Failed to read public key for logging: %v", err)
	}()

	// Check if exists
	_, err1 := os.Stat(privateKeyPath)
	_, err2 := os.Stat(publicKeyPath)
	if err1 == nil && err2 == nil {
		s.logger.LogInfo("Deploy key pair already exists at %s and %s", privateKeyPath, publicKeyPath)
		s.repo.DeployKeys = keys
		return nil
	}

	if err := os.MkdirAll(keyDir, 0700); err != nil {
		s.logger.LogNewError("Failed to create key directory: %v", err)
		return err
	}

	// ------ PRIVATE KEY (PEM format) ------
	privKey, err := ssh.MarshalPrivateKey(priv, deployKeyComment)
	if err != nil {
		s.logger.LogNewError("Failed to marshal private key: %v", err)
		return err
	}

	privPem := pem.EncodeToMemory(privKey)
	if err := os.WriteFile(privateKeyPath, privPem, 0600); err != nil {
		s.logger.LogNewError("Failed to write private key to file: %v", err)
		return err
	}

	// ------ PUBLIC KEY (Authorised Keys format format) ------
	sshPub, err := ssh.NewPublicKey(pub)

	if err != nil {
		s.logger.LogNewError("Failed to generate SSH public key: %v", err)
		return err
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPub)
	pubBytes = bytes.Replace(pubBytes, []byte("\n"), []byte(" "+deployKeyComment+"\n"), 1)
	if err := os.WriteFile(publicKeyPath, pubBytes, 0644); err != nil {
		s.logger.LogNewError("Failed to write public key to file: %v", err)
		return err
	}

	s.repo.DeployKeys = keys
	s.logger.LogInfo("Deploy key pair created at %s and %s", privateKeyPath, publicKeyPath)
	return nil
}
