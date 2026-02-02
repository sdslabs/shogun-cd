package sshclient

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHManager interface {
	// Get an existing SSH client for the target
	GetClient(target string) (*ssh.Client, error)
	// Create a new SSH client, returns it and stores it for later
	NewClient(name string, host string, port int, user string, privateKey []byte) (*ssh.Client, error)
	// Cleanup all SSH clients
	CleanupSSHClients()
}

// Manager is a best-effort SSH client cache for connections
type Manager struct {
	// Use normal mutex as Manager is unique per pipeline and will be access by a single step at a time.
	// Replace this with sync.RWMutex if you implement concurrent step execution in future.
	mu      sync.Mutex
	clients map[string]*ssh.Client
}

func NewSSHManager() SSHManager {
	return &Manager{
		clients: make(map[string]*ssh.Client),
	}
}

func (m *Manager) GetClient(target string) (*ssh.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	client, exists := m.clients[target]
	if !exists {
		return nil, fmt.Errorf("No client available")
	}
	return client, nil
}

func (m *Manager) NewClient(name string, host string, port int, user string, privateKey []byte) (*ssh.Client, error) {
	// It is safe to lock the entire function as sshManager is unique per pipeline
	// and will be accessed by a single step at a time.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new SSH client
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// [TODO]: Implement certificate host verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial SSH: %v", err)
	}
	m.clients[name] = client
	return client, nil
}

func IsDeadClient(client *ssh.Client) bool {
	session, err := client.NewSession()
	if err != nil {
		return true
	}
	defer session.Close()
	return false
}

func (m *Manager) CleanupSSHClients() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, client := range m.clients {
		client.Close()
		delete(m.clients, name)
	}
}
