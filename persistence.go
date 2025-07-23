package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AppData struct {
	RootBullets []*Bullet `json:"rootBullets"`
	Settings    Settings  `json:"settings"`
}

type ConfigManager struct {
	configDir  string
	configFile string
}

func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "ocli")
	configFile := filepath.Join(configDir, "data.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &ConfigManager{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

func (cm *ConfigManager) Save(data *AppData) error {
	// Convert bullets to JSON-serializable format (remove parent references to avoid cycles)
	jsonData := cm.prepareForSerialization(data)
	
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(cm.configFile, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (cm *ConfigManager) Load() (*AppData, error) {
	// Check if config file exists
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		// IMPORTANT: Only create tutorial data for NEW users
		// Existing users' data is always preserved when updating OCLI
		return cm.createDefaultData(), nil
	}

	jsonBytes, err := os.ReadFile(cm.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var data AppData
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Restore parent relationships after loading
	cm.restoreParentRelationships(&data)

	return &data, nil
}

func (cm *ConfigManager) prepareForSerialization(data *AppData) *AppData {
	// Deep copy the data and remove parent references to avoid circular dependencies
	serializedData := &AppData{
		RootBullets: make([]*Bullet, len(data.RootBullets)),
		Settings:    data.Settings,
	}

	for i, bullet := range data.RootBullets {
		serializedData.RootBullets[i] = cm.copyBulletForSerialization(bullet)
	}

	return serializedData
}

func (cm *ConfigManager) copyBulletForSerialization(bullet *Bullet) *Bullet {
	copy := &Bullet{
		ID:        bullet.ID,
		Content:   bullet.Content,
		Children:  make([]*Bullet, len(bullet.Children)),
		Parent:    nil, // Remove parent reference to avoid cycles
		Collapsed: bullet.Collapsed,
		IsEditing: false, // Don't save editing state
		Color:     bullet.Color,
		IsTask:    bullet.IsTask,
		Completed: bullet.Completed,
	}

	for i, child := range bullet.Children {
		copy.Children[i] = cm.copyBulletForSerialization(child)
	}

	return copy
}

func (cm *ConfigManager) restoreParentRelationships(data *AppData) {
	for _, bullet := range data.RootBullets {
		cm.restoreParentRelationshipsRecursive(bullet, nil)
	}
}

func (cm *ConfigManager) restoreParentRelationshipsRecursive(bullet *Bullet, parent *Bullet) {
	bullet.Parent = parent
	bullet.IsEditing = false // Ensure no bullets are in editing state on load

	for _, child := range bullet.Children {
		cm.restoreParentRelationshipsRecursive(child, bullet)
	}
}

func (cm *ConfigManager) createDefaultData() *AppData {
	// Create concise tutorial data
	welcome := NewBullet("Welcome to OCLI!")
	
	// Essential basics
	welcome.AddChild(NewBullet("Press Enter to add bullets, ↑↓ to navigate"))
	welcome.AddChild(NewBullet("Tab/Shift+Tab to indent/outdent"))
	
	// Show task example
	task := NewBullet("Press 't' for tasks, 'x' to complete")
	task.ToggleTask()
	welcome.AddChild(task)
	
	// Show colors
	colored := NewBullet("Press 'c' for colors")
	colored.Color = ColorBlue
	welcome.AddChild(colored)
	
	// Essential features
	collapse := NewBullet("Space to collapse/expand, → to zoom in")
	collapse.AddChild(NewBullet("Hidden content"))
	welcome.AddChild(collapse)
	
	welcome.AddChild(NewBullet("Press 'h' for help, 's' for settings, 'q' to quit"))

	return &AppData{
		RootBullets: []*Bullet{welcome},
		Settings: Settings{
			ShowHierarchyLines: true,
		},
	}
}