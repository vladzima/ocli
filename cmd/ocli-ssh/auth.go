package main

import (
	"bufio"
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// AuthManager handles SSH key authentication for users
type AuthManager struct {
	dataDir        string
	authorizedKeys map[string][]gossh.PublicKey // username -> authorized keys
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(dataDir string) (*AuthManager, error) {
	am := &AuthManager{
		dataDir:        dataDir,
		authorizedKeys: make(map[string][]gossh.PublicKey),
	}

	// Load all user authorized_keys files
	usersDir := filepath.Join(dataDir, "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create users directory: %w", err)
	}

	// Scan for existing users
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read users directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			username := entry.Name()
			if err := am.loadUserKeys(username); err != nil {
				// Log error but continue loading other users
				fmt.Printf("Warning: failed to load keys for user %s: %v\n", username, err)
			}
		}
	}

	return am, nil
}

// loadUserKeys loads authorized keys for a specific user
func (am *AuthManager) loadUserKeys(username string) error {
	keysFile := filepath.Join(am.dataDir, "users", username, "authorized_keys")
	
	// Check if file exists
	if _, err := os.Stat(keysFile); os.IsNotExist(err) {
		// No keys file for this user yet
		return nil
	}

	file, err := os.Open(keysFile)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	var keys []gossh.PublicKey
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			// Skip invalid keys
			continue
		}

		keys = append(keys, key)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading authorized_keys: %w", err)
	}

	am.authorizedKeys[username] = keys
	return nil
}

// Authenticate checks if a public key is authorized for a user
func (am *AuthManager) Authenticate(username string, key ssh.PublicKey) bool {
	userKeys, exists := am.authorizedKeys[username]
	if !exists {
		return false
	}

	// Convert charmbracelet/ssh key to golang.org/x/crypto/ssh key for comparison
	gosshKey, err := convertToGosshKey(key)
	if err != nil {
		return false
	}

	keyData := gosshKey.Marshal()
	for _, authorizedKey := range userKeys {
		if authorizedKey.Type() == gosshKey.Type() && 
			subtle.ConstantTimeCompare(authorizedKey.Marshal(), keyData) == 1 {
			return true
		}
	}

	return false
}

// AuthenticateOrRegister checks auth and optionally registers new users
func (am *AuthManager) AuthenticateOrRegister(username string, key ssh.PublicKey, autoRegister bool) bool {
	// First try normal authentication
	if am.Authenticate(username, key) {
		return true
	}

	// If auto-register is enabled and user doesn't exist, create them
	if autoRegister {
		_, exists := am.authorizedKeys[username]
		if !exists {
			// Auto-register the new user
			if err := am.AddUserKey(username, key); err == nil {
				return true
			}
		}
	}

	return false
}

// AddUserKey adds a new SSH key for a user (used for auto-registration)
func (am *AuthManager) AddUserKey(username string, key ssh.PublicKey) error {
	// Create user directory
	userDir := filepath.Join(am.dataDir, "users", username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	// Convert charmbracelet/ssh key to golang.org/x/crypto/ssh key for storage
	gosshKey, err := convertToGosshKey(key)
	if err != nil {
		return fmt.Errorf("failed to convert key: %w", err)
	}

	// Add to in-memory store
	am.authorizedKeys[username] = append(am.authorizedKeys[username], gosshKey)

	// Write to authorized_keys file
	keysFile := filepath.Join(userDir, "authorized_keys")
	file, err := os.OpenFile(keysFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	// Format the key in authorized_keys format
	keyLine := gossh.MarshalAuthorizedKey(gosshKey)
	if _, err := file.Write(keyLine); err != nil {
		return fmt.Errorf("failed to write key: %w", err)
	}

	return nil
}

// AddUserFromFile adds a user with their public key from a file
func (am *AuthManager) AddUserFromFile(username, publicKeyPath string) error {
	// Read the public key file
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	// Parse the key
	key, _, _, _, err := gossh.ParseAuthorizedKey(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Create user directory
	userDir := filepath.Join(am.dataDir, "users", username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	// Add to in-memory store
	am.authorizedKeys[username] = append(am.authorizedKeys[username], key)

	// Write to authorized_keys file
	keysFile := filepath.Join(userDir, "authorized_keys")
	if err := os.WriteFile(keysFile, keyData, 0600); err != nil {
		return fmt.Errorf("failed to write authorized_keys: %w", err)
	}

	return nil
}

// RemoveUser removes all authentication data for a user
func (am *AuthManager) RemoveUser(username string) error {
	// Remove from memory
	delete(am.authorizedKeys, username)

	// Note: actual directory removal is handled by the server
	return nil
}

// convertToGosshKey converts a charmbracelet/ssh key to golang.org/x/crypto/ssh key
func convertToGosshKey(key ssh.PublicKey) (gossh.PublicKey, error) {
	// Get the key data and type
	keyData := key.Marshal()
	keyType := key.Type()
	
	// Parse it back as a golang.org/x/crypto/ssh key
	gosshKey, err := gossh.ParsePublicKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}
	
	// Verify the type matches
	if gosshKey.Type() != keyType {
		return nil, fmt.Errorf("key type mismatch: expected %s, got %s", keyType, gosshKey.Type())
	}
	
	return gosshKey, nil
}