#!/bin/bash

# Depozit Server Management Script
# This script manages the installation and uninstallation of the depozit-server systemd service.

set -e

# Configuration
USER="depozit"
GROUP="depozit"
BIN_DIR="/opt/depozit"
CONF_DIR="/etc/depozit"
STORAGE_DIR="/var/lib/depozit"
BINARY_NAME="depozit-server"
SERVICE_FILE="/etc/systemd/system/depozit.service"
ENV_FILE="$CONF_DIR/.env"
SERVICE_NAME="depozit"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

install() {
    echo "--- Starting Depozit Server Installation ---"

    # 1. Create System User
    if ! id "$USER" >/dev/null 2>&1; then
        echo "Creating system user: $USER"
        useradd -r -s /bin/false "$USER"
    else
        echo "User $USER already exists."
    fi

    # 2. Prepare Directories
    echo "Preparing directories..."
    mkdir -p "$BIN_DIR"
    mkdir -p "$CONF_DIR"
    mkdir -p "$STORAGE_DIR"

    # 3. Copy Binary (if available in current directory)
    if [ -f "./$BINARY_NAME" ]; then
        echo "Copying binary to $BIN_DIR"
        cp "./$BINARY_NAME" "$BIN_DIR/"
        chmod +x "$BIN_DIR/$BINARY_NAME"
    else
        echo "Warning: $BINARY_NAME not found in current directory. Please copy it to $BIN_DIR later."
    fi

    # 4. Initialize .env file if it doesn't exist
    if [ ! -f "$ENV_FILE" ]; then
        echo "Initializing default .env file in $ENV_FILE"
        cat <<EOF > "$ENV_FILE"
PORT=50051
ROOT_API_KEY=$(openssl rand -hex 32)
STORAGE_PATH=$STORAGE_DIR
CONTAINER=default
EOF
        echo "IMPORTANT: A random ROOT_API_KEY has been generated in $ENV_FILE"
        echo "WARNING: Keep this file secure. Permissions will be set to 600."
    else
        echo ".env file already exists at $ENV_FILE. Skipping initialization."
    fi

    # 5. Set Permissions
    echo "Setting permissions..."
    chown -R "$USER:$GROUP" "$BIN_DIR"
    chown -R "$USER:$GROUP" "$STORAGE_DIR"
    chown -R "$USER:$GROUP" "$CONF_DIR"

    if [ -f "$ENV_FILE" ]; then
        chmod 600 "$ENV_FILE"
        chown "$USER:$GROUP" "$ENV_FILE"
    fi

    chmod 700 "$CONF_DIR"
    chmod 700 "$STORAGE_DIR"

    # 6. Create Systemd Service File
    echo "Creating systemd service file..."
    cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=Depozit Server
After=network.target

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$BIN_DIR
ExecStart=$BIN_DIR/$BINARY_NAME -env $ENV_FILE
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    # 7. Enable and Start Service
    echo "Enabling and starting $SERVICE_NAME service..."
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"

    if [ -f "$BIN_DIR/$BINARY_NAME" ] && [ -f "$ENV_FILE" ]; then
        systemctl start "$SERVICE_NAME"
        echo "Service started successfully."
    else
        echo "Configuration or binary missing. Please ensure both are present and then start the service manually."
    fi

    printf "\n--- Installation Complete ---"
    if [ -f "$ENV_FILE" ]; then
        echo "ROOT_API_KEY: $(grep 'ROOT_API_KEY=' "$ENV_FILE" | cut -d'=' -f2)"
    fi
    echo "Binary location: $BIN_DIR/$BINARY_NAME"
    echo "Config location: $ENV_FILE (Permissions: 600)"
    echo "Storage location: $STORAGE_DIR"
    echo "Check status with: systemctl status $SERVICE_NAME"
}

uninstall() {
    echo "--- Starting Depozit Server Uninstallation ---"

    echo "Stopping and disabling service..."
    systemctl stop "$SERVICE_NAME" || true
    systemctl disable "$SERVICE_NAME" || true

    echo "Killing any remaining processes for user $USER..."
    pkill -u "$USER" || true
    sleep 2

    echo "Removing systemd service file..."
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload

    echo "Removing binary and configuration..."
    rm -rf "$BIN_DIR"
    rm -rf "$CONF_DIR"

    echo "Removing system user: $USER"
    if id "$USER" >/dev/null 2>&1; then
        userdel -f "$USER"
    fi

    echo "--- Uninstallation Complete ---"
}

case "$1" in
    uninstall)
        uninstall
        ;;
    install|*)
        install
        ;;
esac
