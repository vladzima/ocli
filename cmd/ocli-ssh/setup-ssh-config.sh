#!/bin/bash

# OCLI SSH Config Setup
# Configures your SSH to connect to OCLI with just: ssh ocli

set -e

SERVER_HOST="34.61.150.52"
SERVER_PORT="2222"
SSH_CONFIG="$HOME/.ssh/config"

echo "🔧 Setting up SSH config for OCLI..."

# Create .ssh directory if it doesn't exist
mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"

# Backup existing config if it exists
if [ -f "$SSH_CONFIG" ]; then
    cp "$SSH_CONFIG" "$SSH_CONFIG.backup.$(date +%s)"
    echo "📋 Backed up existing SSH config"
fi

# Check if OCLI config already exists
if grep -q "Host ocli" "$SSH_CONFIG" 2>/dev/null; then
    echo "⚠️  OCLI config already exists in SSH config"
    echo "🔄 Updating existing configuration..."
    # Remove existing ocli config block
    sed -i.tmp '/^Host ocli$/,/^$/d' "$SSH_CONFIG" 2>/dev/null || true
    rm -f "$SSH_CONFIG.tmp"
fi

# Add OCLI SSH configuration
cat >> "$SSH_CONFIG" << EOF

# OCLI SSH Server Configuration
Host ocli
    HostName $SERVER_HOST
    Port $SERVER_PORT
    User $(whoami)
    ServerAliveInterval 60
    ServerAliveCountMax 3

EOF

echo "✅ SSH config updated successfully!"
echo ""
echo "📋 Usage:"
echo "  ssh ocli                # Connect as $(whoami)"  
echo "  ssh username@ocli       # Connect as specific user"
echo ""
echo "🎯 Quick start:"
echo "  Just run: ssh ocli"
echo ""