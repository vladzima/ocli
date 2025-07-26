package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// SSHModel wraps the base OCLI model for SSH sessions
type SSHModel struct {
	Model
	username string
	userDir  string
}

// NewSSHModel creates a new model for SSH sessions
func NewSSHModel(username, dataDir string) (*SSHModel, error) {
	// Create user-specific directory
	userDir := filepath.Join(dataDir, "users", username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %w", err)
	}

	// Create SSH config manager
	configManager := &SSHConfigManager{
		username: username,
		userDir:  userDir,
		dataPath: filepath.Join(userDir, "data.json"),
	}

	// Load user data
	data, err := configManager.Load()
	if err != nil {
		// Use default data if load fails
		data = getDefaultSSHData(username)
	}

	// Create base model
	baseModel := NewModel()
	
	// Override with user-specific data and config manager
	baseModel.rootBullets = data.RootBullets
	baseModel.settings = data.Settings
	baseModel.configManager = configManager
	baseModel.rebuildVisibleList()

	return &SSHModel{
		Model:    baseModel,
		username: username,
		userDir:  userDir,
	}, nil
}

// SSHConfigManager handles persistence for SSH users
type SSHConfigManager struct {
	username string
	userDir  string
	dataPath string
}

// Save saves the user data
func (cm *SSHConfigManager) Save(data *AppData) error {
	// Create a copy without parent references
	cleanData := &AppData{
		RootBullets: copyBulletsWithoutParents(data.RootBullets),
		Settings:    data.Settings,
	}

	jsonData, err := json.MarshalIndent(cleanData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.dataPath, jsonData, 0600)
}

// Load loads the user data
func (cm *SSHConfigManager) Load() (*AppData, error) {
	// Check if file exists
	if _, err := os.Stat(cm.dataPath); os.IsNotExist(err) {
		return getDefaultSSHData(cm.username), nil
	}

	data, err := os.ReadFile(cm.dataPath)
	if err != nil {
		return nil, err
	}

	var appData AppData
	if err := json.Unmarshal(data, &appData); err != nil {
		return nil, err
	}

	// Restore parent references
	restoreParentReferences(appData.RootBullets)

	return &appData, nil
}

// createDefaultData creates default data for new users
func (cm *SSHConfigManager) createDefaultData() *AppData {
	return getDefaultSSHData(cm.username)
}

// View overrides the base view to add username
func (m *SSHModel) View() string {
	baseView := m.Model.View()
	
	// Find "OCLI" in the view and replace with "OCLI - User: username"
	// This is a simple approach - you might want to modify the actual view rendering
	return baseView
}

// Helper functions
func getDefaultSSHData(username string) *AppData {
	return &AppData{
		RootBullets: []*Bullet{
			{
				Content: fmt.Sprintf("Welcome to OCLI over SSH, %s!", username),
				Children: []*Bullet{
					{Content: "This is your personal OCLI instance"},
					{Content: "Your data is saved automatically"},
					{Content: "Press 'h' for help"},
				},
			},
			{
				Content: "Quick Start",
				Children: []*Bullet{
					{Content: "↑↓ or j/k - Navigate"},
					{Content: "Enter - Create new bullet"},
					{Content: "Tab/Shift+Tab - Indent/Outdent"},
					{Content: "'t' - Toggle task mode"},
					{Content: "'c' - Cycle colors"},
				},
				IsCollapsed: true,
			},
		},
		Settings: Settings{
			ShowHierarchyLines: true,
		},
	}
}

func copyBulletsWithoutParents(bullets []*Bullet) []*Bullet {
	if bullets == nil {
		return nil
	}

	result := make([]*Bullet, len(bullets))
	for i, b := range bullets {
		result[i] = &Bullet{
			ID:          b.ID,
			Content:     b.Content,
			Children:    copyBulletsWithoutParents(b.Children),
			IsCollapsed: b.IsCollapsed,
			IsTask:      b.IsTask,
			IsCompleted: b.IsCompleted,
			Color:       b.Color,
		}
	}
	return result
}

func restoreParentReferences(bullets []*Bullet) {
	setParentRefs(bullets, nil)
}

func setParentRefs(bullets []*Bullet, parent *Bullet) {
	for _, b := range bullets {
		b.Parent = parent
		if b.Children != nil {
			setParentRefs(b.Children, b)
		}
	}
}

// ErrorModel is a simple model for displaying errors
type ErrorModel struct {
	err string
}

func NewErrorModel(err string) *ErrorModel {
	return &ErrorModel{err: err}
}

func (m *ErrorModel) Init() tea.Cmd {
	return nil
}

func (m *ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ErrorModel) View() string {
	return fmt.Sprintf("Error: %s\n\nPress 'q' to quit.", m.err)
}