package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "2222"
)

func main() {
	// Check environment variables first (for Railway/cloud deployments)
	envPort := os.Getenv("PORT")
	if envPort == "" {
		envPort = os.Getenv("OCLI_SSH_PORT")
		if envPort == "" {
			envPort = defaultPort
		}
	}

	envHost := os.Getenv("OCLI_SSH_HOST")
	if envHost == "" {
		envHost = defaultHost
	}

	envDataDir := os.Getenv("OCLI_SSH_DATA_DIR")
	if envDataDir == "" {
		// Use a user-writable directory by default
		if homeDir, err := os.UserHomeDir(); err == nil {
			envDataDir = homeDir + "/.ocli-ssh"
		} else {
			envDataDir = "./data"
		}
	}

	envAutoRegister := false
	if ar := os.Getenv("OCLI_SSH_AUTO_REGISTER"); ar != "" {
		envAutoRegister, _ = strconv.ParseBool(ar)
	}

	var (
		host         = flag.String("host", envHost, "Host to bind SSH server to")
		port         = flag.String("port", envPort, "Port to bind SSH server to")
		dataDir      = flag.String("data-dir", envDataDir, "Directory to store user data")
		keyPath      = flag.String("key", "", "Path to SSH host key (generates if not specified)")
		addUser      = flag.String("add-user", "", "Add a new user (format: username:path/to/public_key.pub)")
		delUser      = flag.String("del-user", "", "Remove a user")
		autoRegister = flag.Bool("auto-register", envAutoRegister, "Automatically register new users on first connection")
	)
	flag.Parse()

	// Handle user management commands
	if *addUser != "" {
		if err := handleAddUser(*dataDir, *addUser); err != nil {
			log.Fatal("Failed to add user:", err)
		}
		return
	}

	if *delUser != "" {
		if err := handleDelUser(*dataDir, *delUser); err != nil {
			log.Fatal("Failed to delete user:", err)
		}
		return
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize server
	srv, err := NewServer(*host, *port, *dataDir, *keyPath, *autoRegister)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting OCLI SSH server on %s:%s", *host, *port)
	log.Printf("Data directory: %s", *dataDir)
	log.Printf("SSH key path: %s", *keyPath)
	if *autoRegister {
		log.Println("Auto-registration: ENABLED (new users will be created automatically)")
	} else {
		log.Println("Auto-registration: DISABLED")
	}
	log.Println("")
	if !*autoRegister {
		log.Println("To add users: ocli-ssh --add-user username:path/to/key.pub")
	}
	log.Println("To connect: ssh username@hostname -p", *port)
	
	// Debug: Check if we can write to data directory
	testFile := *dataDir + "/test"
	if f, err := os.Create(testFile); err != nil {
		log.Printf("WARNING: Cannot write to data directory: %v", err)
	} else {
		f.Close()
		os.Remove(testFile)
		log.Printf("Data directory is writable")
	}
	
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-done
	log.Println("Shutting down SSH server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Failed to shutdown server:", err)
	}
}

func handleAddUser(dataDir, userSpec string) error {
	// Parse username:keyfile format
	parts := strings.SplitN(userSpec, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use: username:path/to/public_key.pub")
	}

	username := parts[0]
	keyPath := parts[1]

	// Create auth manager
	authManager, err := NewAuthManager(dataDir)
	if err != nil {
		return err
	}

	// Add user
	if err := authManager.AddUserFromFile(username, keyPath); err != nil {
		return err
	}

	fmt.Printf("User '%s' added successfully\n", username)
	return nil
}

func handleDelUser(dataDir, username string) error {
	// Create server just to remove user
	srv := &Server{dataDir: dataDir}
	
	if err := srv.RemoveUser(username); err != nil {
		return err
	}

	fmt.Printf("User '%s' removed successfully\n", username)
	return nil
}