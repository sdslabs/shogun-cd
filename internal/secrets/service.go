package secrets

import (
	"fmt"
	"regexp"
	"sync"
)

type SecretService interface {
	// Fetches the secret value if present othervise error
	FetchSecret(name string) (string, error)
	// Adds a new secret with the given name and value, should take care of encryption internally
	AddSecret(name string, value string)
	// Deletes the secret with the given name
	DeleteSecret(name string)
	// [TODO] : Currently add is used to update as well
	// Discuss about having a separate update method
	// Resolve secret value in a string with {{NAME}} patterns
	ResolveSecrets(input string) string
}

// [TODO]: This is a temporary in-memory implementation for testing,
// we will use postgress to store encrypted values or use services like HashiCorp Vault, AWS Secrets Manager, etc.
type Service struct {
	SecretMap map[string]string
	mu        sync.RWMutex
}

func NewSecretService() SecretService {
	s := Service{
		SecretMap: make(map[string]string),
	}
	return &s
}

func (s *Service) FetchSecret(name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.SecretMap[name]
	if !exists {
		return "", fmt.Errorf("Secret value not defined")
	}
	return value, nil
}

func (s *Service) AddSecret(name string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SecretMap[name] = value
}

func (s *Service) DeleteSecret(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.SecretMap, name)
}

func (s *Service) ResolveSecrets(input string) string {
	var secretRegex = regexp.MustCompile(`\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`)
	return secretRegex.ReplaceAllStringFunc(input, func(match string) string {
		s.mu.Lock()
		defer s.mu.Unlock()
		secretName := secretRegex.FindStringSubmatch(match)[1]
		val, exists := s.SecretMap[secretName]
		if exists {
			return val
		}
		return ""
	})
}
