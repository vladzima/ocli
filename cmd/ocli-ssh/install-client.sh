#!/bin/bash

# OCLI SSH Client Installer
# This script creates a convenient 'ocli' command for connecting to your OCLI SSH server

set -e

SERVER_HOST="34.61.150.52"
SERVER_PORT="2222"
SCRIPT_NAME="ocli"

echo "🚀 Installing OCLI SSH Client..."

# Determine install location
if [[ ":$PATH:" == *":/usr/local/bin:"* ]] && [[ -w "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
elif [[ ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo "⚠️  Adding $INSTALL_DIR to your PATH..."
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc 2>/dev/null || true
fi

# Create the ocli script content
OCLI_SCRIPT='#!/bin/bash

# OCLI SSH Client
# Connect to your personal OCLI instance

SERVER_HOST="34.61.150.52"
SERVER_PORT="2222"

# Get username from arguments or system
if [ $# -eq 1 ]; then
    USERNAME="$1"
else
    USERNAME=$(whoami)
fi

echo "🔗 Connecting to OCLI as $USERNAME..."
echo "📝 Your personal notes are saved on the server"
echo ""

# Connect to OCLI SSH server
exec ssh "$USERNAME@$SERVER_HOST" -p "$SERVER_PORT"'

# Write the script
if [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
    echo "$OCLI_SCRIPT" | sudo tee "$INSTALL_DIR/$SCRIPT_NAME" > /dev/null
    sudo chmod +x "$INSTALL_DIR/$SCRIPT_NAME"
else
    echo "$OCLI_SCRIPT" > "$INSTALL_DIR/$SCRIPT_NAME"
    chmod +x "$INSTALL_DIR/$SCRIPT_NAME"
fi

echo "✅ OCLI client installed successfully!"
echo ""
echo "📋 Usage:"
echo "  ocli                    # Connect as $(whoami)"
echo "  ocli username           # Connect as specific user"
echo ""
echo "🎯 Quick start:"
echo "  Just run: ocli"
echo ""

# Test if command is available
if command -v ocli >/dev/null 2>&1; then
    echo "✨ Ready to use! Type 'ocli' to connect."
else
    echo "⚠️  You may need to restart your terminal or run:"
    echo "    source ~/.bashrc"
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
fi