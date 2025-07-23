.PHONY: build install clean test run

# Binary name
BINARY_NAME=ocli

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

# Install the binary to GOPATH/bin
install:
	$(GOCMD) install .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests (if any exist)
test:
	$(GOTEST) -v ./...

# Run the application directly
run:
	$(GOCMD) run .

# Build for multiple platforms
build-all:
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o builds/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o builds/$(BINARY_NAME)-linux-arm64 .
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o builds/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o builds/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o builds/$(BINARY_NAME)-windows-amd64.exe .

# Tidy up dependencies
tidy:
	$(GOMOD) tidy

# Create builds directory
prepare-builds:
	mkdir -p builds

# Full release build
release: prepare-builds tidy build-all
	@echo "Release builds created in builds/ directory"