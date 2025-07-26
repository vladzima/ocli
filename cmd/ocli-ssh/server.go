package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

type Server struct {
	wishServer   *ssh.Server
	dataDir      string
	authManager  *AuthManager
	autoRegister bool
}

func NewServer(host, port, dataDir, keyPath string, autoRegister bool) (*Server, error) {
	// Create auth manager
	authManager, err := NewAuthManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}
	
	// Debug logging
	fmt.Printf("Creating SSH server with host=%s, port=%s, dataDir=%s, keyPath=%s\n", host, port, dataDir, keyPath)

	s := &Server{
		dataDir:      dataDir,
		authManager:  authManager,
		autoRegister: autoRegister,
	}

	// Set up middleware
	middleware := []wish.Middleware{
		bm.Middleware(s.teaHandler),
		logging.Middleware(),
	}

	// Pre-generate SSH key if none specified (Railway might have issues with auto-generation)
	if keyPath == "" {
		keyPath = filepath.Join(dataDir, "ssh_host_key")
		if err := generateSSHKey(keyPath); err != nil {
			fmt.Printf("Failed to generate SSH key: %v\n", err)
			return nil, fmt.Errorf("failed to generate SSH key: %w", err)
		}
		fmt.Printf("Generated SSH key at: %s\n", keyPath)
	}

	// Configure server options
	opts := []ssh.Option{
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithMiddleware(middleware...),
		wish.WithPublicKeyAuth(s.authHandler),
		wish.WithHostKeyPath(keyPath),
	}

	fmt.Printf("Creating wish server with %d options\n", len(opts))
	fmt.Printf("Server address: %s:%s\n", host, port)
	fmt.Printf("SSH key path: %s\n", keyPath)
	
	srv, err := wish.NewServer(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create wish server: %w", err)
	}
	fmt.Printf("Wish server created successfully\n")

	s.wishServer = srv
	return s, nil
}

func (s *Server) Start() error {
	fmt.Printf("Starting SSH server...\n")
	
	// Test if we can bind to the address first
	addr := s.wishServer.Addr
	fmt.Printf("Attempting to bind to address: %s\n", addr)
	
	err := s.wishServer.ListenAndServe()
	if err != nil {
		fmt.Printf("SSH server failed to start: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
		
		// Check if it's a binding error
		if opErr, ok := err.(*net.OpError); ok {
			fmt.Printf("Network operation error: %s\n", opErr.Op)
			fmt.Printf("Network: %s\n", opErr.Net)
			fmt.Printf("Source: %v\n", opErr.Source)
			fmt.Printf("Addr: %v\n", opErr.Addr)
			if opErr.Err != nil {
				fmt.Printf("Underlying error: %v\n", opErr.Err)
			}
		}
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.wishServer.Shutdown(ctx)
}

func (s *Server) teaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	// Get username from SSH session
	username := sess.User()
	if username == "" {
		// This shouldn't happen with proper auth, but handle it gracefully
		username = "anonymous"
	}

	// Create user-specific model
	model, err := NewSSHModel(username, s.dataDir)
	if err != nil {
		// Return error model
		return NewErrorModel(fmt.Sprintf("Failed to initialize: %v", err)), []tea.ProgramOption{tea.WithAltScreen()}
	}

	// Return model with appropriate options
	return model, []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
}

func (s *Server) authHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	// Get username from context
	username := ctx.User()
	if username == "" {
		return false
	}

	// Check if key is authorized for this user (with auto-registration if enabled)
	return s.authManager.AuthenticateOrRegister(username, key, s.autoRegister)
}

// AddUser adds a new user with their SSH public key
func (s *Server) AddUser(username string, publicKeyPath string) error {
	return s.authManager.AddUserFromFile(username, publicKeyPath)
}

// RemoveUser removes a user and all their data
func (s *Server) RemoveUser(username string) error {
	// Remove authentication
	if err := s.authManager.RemoveUser(username); err != nil {
		return err
	}

	// Remove user data directory
	userDir := filepath.Join(s.dataDir, "users", username)
	return os.RemoveAll(userDir)
}

// generateSSHKey generates an ED25519 SSH host key
func generateSSHKey(keyPath string) error {
	fmt.Printf("Generating SSH key at: %s\n", keyPath)
	
	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Printf("SSH key already exists at: %s\n", keyPath)
		return nil
	}
	
	// Generate ED25519 key pair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ED25519 key: %w", err)
	}

	// Convert to PKCS8 format
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	}

	// Write to file
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("failed to write PEM data: %w", err)
	}

	fmt.Printf("Successfully generated SSH key\n")
	return nil
}