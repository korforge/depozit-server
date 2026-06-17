#!/bin/bash

# Depozit Server Installer
# Installs without root. Requires: systemd (user session), openssl.

set -e

BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.depozit"
STORAGE_DIR="$HOME/.depozit/storage"
SYSTEMD_DIR="$HOME/.config/systemd/user"
ENV_FILE="$CONFIG_DIR/.env"
SERVICE_FILE="$SYSTEMD_DIR/depozit.service"
SERVICE_NAME="depozit"
SERVER_BINARY="depozit-server"
CLI_BINARY="depozit"

# Resolve script directory so install can be run from anywhere
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

install() {
    echo "--- Installing Depozit ---"

    # Check for the compiled server binary
    if [ ! -f "$SCRIPT_DIR/$SERVER_BINARY" ]; then
        echo "Error: $SERVER_BINARY not found in $SCRIPT_DIR"
        echo "Build it first with: go build -o $SERVER_BINARY ./src"
        exit 1
    fi

    # Check for the CLI binary
    if [ ! -f "$SCRIPT_DIR/$CLI_BINARY" ]; then
        echo "Error: $CLI_BINARY not found in $SCRIPT_DIR"
        exit 1
    fi

    # Check systemd user session is available
    if ! systemctl --user status >/dev/null 2>&1; then
        echo "Error: systemd user session is not available."
        echo "Make sure you are logged in as a regular user with a running session."
        exit 1
    fi

    echo "Creating directories..."
    mkdir -p "$BIN_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$STORAGE_DIR"
    mkdir -p "$SYSTEMD_DIR"

    echo "Installing binaries to $BIN_DIR..."
    cp "$SCRIPT_DIR/$SERVER_BINARY" "$BIN_DIR/$SERVER_BINARY"
    cp "$SCRIPT_DIR/$CLI_BINARY" "$BIN_DIR/$CLI_BINARY"
    chmod +x "$BIN_DIR/$SERVER_BINARY"
    chmod +x "$BIN_DIR/$CLI_BINARY"

    if [ ! -f "$ENV_FILE" ]; then
        echo "Creating default config at $ENV_FILE..."
        cat > "$ENV_FILE" <<EOF
PORT=50051
ROOT_API_KEY=$(openssl rand -hex 32)
STORAGE_PATH=$STORAGE_DIR
CONTAINER=default
EOF
        chmod 600 "$ENV_FILE"
        echo "IMPORTANT: A random ROOT_API_KEY has been generated."
    else
        echo "Config already exists at $ENV_FILE, skipping."
    fi

    echo "Creating systemd user service..."
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Depozit Server
After=network.target

[Service]
Type=simple
ExecStart=$BIN_DIR/$SERVER_BINARY -env $ENV_FILE
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable "$SERVICE_NAME"
    systemctl --user start "$SERVICE_NAME"

    # Add ~/.local/bin to PATH in shell rc files if not already present
    _ensure_path

    echo ""
    echo "--- Installation Complete ---"
    echo "ROOT_API_KEY: $(grep 'ROOT_API_KEY=' "$ENV_FILE" | cut -d'=' -f2)"
    echo "Port:         $(grep 'PORT=' "$ENV_FILE" | cut -d'=' -f2)"
    echo "Storage:      $STORAGE_DIR"
    echo "Config:       $ENV_FILE"
    echo ""
    echo "Manage with: depozit {start|stop|status|uninstall}"
    echo "(You may need to open a new terminal for 'depozit' to be on PATH.)"
}

_ensure_path() {
    local line='export PATH="$HOME/.local/bin:$PATH"'
    for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [ -f "$rc" ] && ! grep -qF '.local/bin' "$rc"; then
            echo "" >> "$rc"
            echo "# Added by depozit installer" >> "$rc"
            echo "$line" >> "$rc"
            echo "Added $BIN_DIR to PATH in $rc"
        fi
    done
}

install
