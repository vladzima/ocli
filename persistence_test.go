package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDataPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ocli_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config manager with temp directory
	configFile := filepath.Join(tempDir, "data.json")
	cm := &ConfigManager{
		configDir:  tempDir,
		configFile: configFile,
	}

	// Test 1: New installation should get tutorial data
	data, err := cm.Load()
	if err != nil {
		t.Fatalf("Failed to load default data: %v", err)
	}

	if len(data.RootBullets) == 0 {
		t.Fatal("Expected tutorial data for new installation")
	}

	if data.RootBullets[0].Content != "Welcome to OCLI!" {
		t.Errorf("Expected tutorial welcome message, got: %s", data.RootBullets[0].Content)
	}

	// Test 2: Save user data
	userBullet := NewBullet("My important work data")
	userData := &AppData{
		RootBullets: []*Bullet{userBullet},
		Settings:    Settings{ShowHierarchyLines: false},
	}

	err = cm.Save(userData)
	if err != nil {
		t.Fatalf("Failed to save user data: %v", err)
	}

	// Test 3: Load should return user data, NOT tutorial
	loadedData, err := cm.Load()
	if err != nil {
		t.Fatalf("Failed to load user data: %v", err)
	}

	if len(loadedData.RootBullets) != 1 {
		t.Fatalf("Expected 1 bullet, got %d", len(loadedData.RootBullets))
	}

	if loadedData.RootBullets[0].Content != "My important work data" {
		t.Errorf("Expected user data, got tutorial: %s", loadedData.RootBullets[0].Content)
	}

	if loadedData.Settings.ShowHierarchyLines != false {
		t.Error("Expected user settings to be preserved")
	}

	// Test 4: Verify file actually contains user data
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var fileJson AppData
	err = json.Unmarshal(fileData, &fileJson)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if fileJson.RootBullets[0].Content != "My important work data" {
		t.Error("User data not properly saved to file")
	}
}

func TestDataPreservationOnUpdate(t *testing.T) {
	// Simulate updating OCLI with existing user data
	tempDir, err := os.MkdirTemp("", "ocli_update_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "data.json")

	// Step 1: Create existing user data (simulating old version)
	existingData := &AppData{
		RootBullets: []*Bullet{
			{
				ID:      "user-1",
				Content: "Important project notes",
				Color:   ColorRed,
				IsTask:  true,
			},
			{
				ID:      "user-2", 
				Content: "Meeting agenda",
				Color:   ColorBlue,
			},
		},
		Settings: Settings{ShowHierarchyLines: false},
	}

	// Save existing data to file
	jsonBytes, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal existing data: %v", err)
	}

	err = os.WriteFile(configFile, jsonBytes, 0644)
	if err != nil {
		t.Fatalf("Failed to write existing data: %v", err)
	}

	// Step 2: Simulate OCLI update (new ConfigManager loading existing data)
	cm := &ConfigManager{
		configDir:  tempDir,
		configFile: configFile,
	}

	loadedData, err := cm.Load()
	if err != nil {
		t.Fatalf("Failed to load after update: %v", err)
	}

	// Step 3: Verify all user data is preserved
	if len(loadedData.RootBullets) != 2 {
		t.Fatalf("Expected 2 user bullets, got %d", len(loadedData.RootBullets))
	}

	if loadedData.RootBullets[0].Content != "Important project notes" {
		t.Errorf("User content not preserved: %s", loadedData.RootBullets[0].Content)
	}

	if loadedData.RootBullets[0].Color != ColorRed {
		t.Error("User bullet color not preserved")
	}

	if !loadedData.RootBullets[0].IsTask {
		t.Error("User task status not preserved")
	}

	if loadedData.Settings.ShowHierarchyLines != false {
		t.Error("User settings not preserved")
	}

	// Step 4: Verify NO tutorial data was added
	for _, bullet := range loadedData.RootBullets {
		if bullet.Content == "Welcome to OCLI!" {
			t.Error("Tutorial data should not appear for existing users")
		}
	}
}