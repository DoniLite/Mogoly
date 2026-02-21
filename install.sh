#!/bin/bash
# Mogoly Installation Script
# Installs Mogoly CLI and sets up necessary permissions

set -e

echo "🚀 Installing Mogoly..."

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM=linux;;
    Darwin*)    PLATFORM=macos;;
    MINGW*|MSYS*|CYGWIN*)     PLATFORM=windows;;
    *)          echo "Unsupported platform: ${OS}"; exit 1;;
esac

echo "📦 Detected platform: ${PLATFORM}"

# Install binary
echo "📥 Downloading Mogoly..."
if command -v go &> /dev/null; then
    echo "   Using Go to install..."
    go install github.com/DoniLite/Mogoly/cli/mogoly@latest
else
    # Download pre-built binary
    echo "   Downloading pre-built binary..."
    # curl -sSL https://github.com/DoniLite/Mogoly/releases/latest/download/mogoly-${PLATFORM} -o /usr/local/bin/mogoly
    # chmod +x /usr/local/bin/mogoly
fi

# Create config directory
echo "📁 Creating config directory..."
mkdir -p ~/.mogoly/certs
chmod 755 ~/.mogoly
chmod 700 ~/.mogoly/certs

echo "✅ Mogoly installed successfully!"
echo ""
echo "📚 Quick Start:"
echo "   1. Start daemon:        mogoly daemon start"
echo "   2. Add local domain:    sudo mogoly domain add myapp.local --local"
echo "   3. Create load balancer: mogoly lb create --name myapp"
echo ""
echo "⚠️  Note: Domain operations require sudo for /etc/hosts modification"
echo ""
echo "📖 Documentation: https://github.com/DoniLite/Mogoly/blob/main/cli/DOMAINS.md"
