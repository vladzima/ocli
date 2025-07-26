package main

import (
	"context"
	"fmt"
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

	// Configure server options
	opts := []wish.Option{
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithMiddleware(middleware...),
		wish.WithPublicKeyAuth(s.authHandler),
	}

	// Add host key path if specified
	if keyPath != "" {
		opts = append(opts, wish.WithHostKeyPath(keyPath))
	}

	srv, err := wish.NewServer(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create wish server: %w", err)
	}

	s.wishServer = srv
	return s, nil
}

func (s *Server) Start() error {
	return s.wishServer.ListenAndServe()
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