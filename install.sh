#!/usr/bin/env bash

# point-tui installation and uninstallation script
# Usage: 
#   Install:   curl -sSL https://raw.githubusercontent.com/dariy/point-tui/main/install.sh | bash
#   Uninstall: curl -sSL https://raw.githubusercontent.com/dariy/point-tui/main/install.sh | bash -s -- uninstall

set -euo pipefail

BINARY_NAME="point-tui"
REPO_URL="https://github.com/dariy/point-tui.git"

# Potential installation directories
LOCAL_BIN="$HOME/.local/bin"
SYSTEM_BIN="/usr/local/bin"

# Determine the best install directory
if [[ ":$PATH:" == *":$LOCAL_BIN:"* ]]; then
    INSTALL_DIR="$LOCAL_BIN"
else
    INSTALL_DIR="$SYSTEM_BIN"
fi

# --- Uninstallation Logic ---
if [[ "${1:-}" == "uninstall" ]]; then
    echo "Uninstalling $BINARY_NAME..."
    
    UNINSTALLED=false
    for dir in "$LOCAL_BIN" "$SYSTEM_BIN"; do
        if [[ -f "$dir/$BINARY_NAME" ]]; then
            echo "Removing $dir/$BINARY_NAME"
            if [[ -w "$dir" ]]; then
                if rm "$dir/$BINARY_NAME"; then
                    UNINSTALLED=true
                else
                    echo "Error: Failed to remove $dir/$BINARY_NAME"
                fi
            else
                if sudo rm "$dir/$BINARY_NAME"; then
                    UNINSTALLED=true
                else
                    echo "Error: Failed to remove $dir/$BINARY_NAME with sudo"
                fi
            fi
        fi
    done

    if [ "$UNINSTALLED" = true ]; then
        echo "Successfully uninstalled $BINARY_NAME."
    else
        echo "$BINARY_NAME not found in common installation directories."
    fi
    exit 0
fi

# --- Installation Logic ---

echo "Starting installation of $BINARY_NAME..."

# 1. Environment Checks
if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' is not installed. Please install Go 1.21 or later."
    exit 1
fi

if ! command -v git >/dev/null 2>&1; then
    echo "Error: 'git' is not installed. Please install git."
    exit 1
fi

if ! command -v ffmpeg >/dev/null 2>&1; then
    echo "Warning: 'ffmpeg' not found. Video previews will not be available."
fi

# 2. Prepare Install Directory
if [[ ! -d "$INSTALL_DIR" ]]; then
    echo "Creating directory $INSTALL_DIR..."
    if [[ "$INSTALL_DIR" == "$HOME"* ]]; then
        mkdir -p "$INSTALL_DIR"
    else
        sudo mkdir -p "$INSTALL_DIR"
    fi
fi

# 3. Build from Source
TMP_DIR=$(mktemp -d)
echo "Cloning repository to $TMP_DIR..."
git clone --depth 1 "$REPO_URL" "$TMP_DIR"

pushd "$TMP_DIR" >/dev/null
echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME" ./cmd/point-tui
popd >/dev/null

# 4. Install Binary
echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
if [[ -w "$INSTALL_DIR" ]]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# 5. Cleanup
rm -rf "$TMP_DIR"

echo "--------------------------------------------------"
echo "Successfully installed $BINARY_NAME to $INSTALL_DIR!"
echo ""
echo "To run the application:"
echo "  $BINARY_NAME"
echo ""
echo "To uninstall:"
echo "  curl -sSL https://raw.githubusercontent.com/dariy/point-tui/main/install.sh | bash -s -- uninstall"
echo "--------------------------------------------------"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Warning: $INSTALL_DIR is not in your PATH."
    echo "You may need to add it to your shell configuration (e.g., .bashrc or .zshrc):"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi
