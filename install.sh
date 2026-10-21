#!/bin/bash

# Depozit Server Installer
# Installs without root. Requires: systemd (user session), openssl, and curl + sha256sum
# when downloading the binaries.
#
#   curl -fsSL https://raw.githubusercontent.com/korforge/depozit-server/main/install.sh | bash
#
# Run from a clone or a downloaded tag archive, it installs the files next to it instead.

set -e

REPO="${DEPOZIT_REPO:-korforge/depozit-server}"

# Branch downloaded when the script runs without the repo files next to it, as with curl | bash.
# A piped script cannot see its own URL, so the pre-push hook sets this to the pushed branch:
# the install.sh on each branch downloads the binaries of that branch
BRANCH="feat/streaming"
BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.depozit"
DEFAULT_STORAGE_DIR="$HOME/.depozit/storage"
SYSTEMD_DIR="$HOME/.config/systemd/user"
ENV_FILE="$CONFIG_DIR/.env"
DEFAULT_KEYS_FILE="$CONFIG_DIR/keys.env"
SERVICE_FILE="$SYSTEMD_DIR/depozit.service"
SERVICE_NAME="depozit"
SERVER_BINARY="depozit-server"
CLI_BINARY="depozit"

usage() {
    cat <<EOF
Usage: install.sh [options]

Options:
  --host HOST         Bind address (default 127.0.0.1, only this machine can connect)
  --port PORT         gRPC port (default 50051)
  --storage DIR       Directory where files are stored (default $DEFAULT_STORAGE_DIR)
  --keys-file FILE    File with ROOT_API_KEY and ENCRYPTION_KEY (default $DEFAULT_KEYS_FILE).
                      Generated with mode 600 when it does not exist, used as is when it does
  -h, --help          Show this help

Never pass ROOT_API_KEY or ENCRYPTION_KEY as arguments: they would show up in ps and
shell history. Let the installer generate the keys file, or write it yourself with mode 600.
EOF
}

# Sets KEY=VALUE in the config, replacing any earlier line for KEY
_set_env() {
    sed -i "/^$1=/d" "$ENV_FILE"
    printf '%s=%s\n' "$1" "$2" >> "$ENV_FILE"
}

# Reads KEY from the config, or from the file given as $2
_read_env() {
    grep "^$1=" "${2:-$ENV_FILE}" 2>/dev/null | tail -n 1 | cut -d'=' -f2- | tr -d "\"'" || true
}

_fail() {
    echo "Error: $*" >&2
    exit 1
}

# Gets the server built by "make build" and the depozit script into $1: from the repo files next to
# this script when there are any, otherwise from the $BRANCH branch. Then checks them against
# dist/sha256sums.txt
_fetch() {
    local dir="$1" arch root

    command -v sha256sum >/dev/null || _fail "sha256sum is required"

    case "$(uname -m)" in
        x86_64 | amd64) arch="amd64" ;;
        aarch64 | arm64) arch="arm64" ;;
        *) _fail "unsupported architecture $(uname -m)" ;;
    esac

    local local_root=""

    # BASH_SOURCE is empty when the script comes from a pipe, so there is nothing next to it
    if [ -f "${BASH_SOURCE[0]}" ]; then
        local_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        [ -f "$local_root/dist/sha256sums.txt" ] || local_root=""
    fi

    if [ -n "$local_root" ]; then
        root="$local_root"
        echo "Using files from $root for linux/$arch..."
    else
        command -v curl >/dev/null || _fail "curl is required"
        root="https://raw.githubusercontent.com/$REPO/$BRANCH"
        echo "Downloading $BRANCH for linux/$arch..."
    fi

    # Paths relative to the repo root, as listed in dist/sha256sums.txt
    local server="dist/$SERVER_BINARY-linux-$arch" cli="scripts/$CLI_BINARY" path

    for path in "$server" "$cli" dist/sha256sums.txt; do
        if [ -n "$local_root" ]; then
            cp "$root/$path" "$dir/$(basename "$path")" 2>/dev/null || _fail "$root/$path not found. Build it first with: make build"
        else
            curl -fsSL --proto '=https' --tlsv1.2 -o "$dir/$(basename "$path")" "$root/$path" ||
                _fail "could not download $root/$path"
        fi
    done

    local expected actual

    for path in "$server" "$cli"; do
        expected="$(awk -v f="$path" '$2 == f || $2 == "*" f { print $1 }' "$dir/sha256sums.txt")"
        actual="$(sha256sum "$dir/$(basename "$path")" | cut -d' ' -f1)"

        if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
            _fail "checksum mismatch for $path, the files are corrupted or out of date. In a clone, run: make build"
        fi
    done

    mv "$dir/$SERVER_BINARY-linux-$arch" "$dir/$SERVER_BINARY"
}

# Creates the config, or updates an existing one
_write_config() {
    if [ ! -f "$ENV_FILE" ]; then
        echo "Creating config at $ENV_FILE..."
        (umask 077 && : > "$ENV_FILE")
        _set_env HOST 127.0.0.1
    else
        echo "Config already exists at $ENV_FILE, keeping it."

        # Older servers defaulted to 0.0.0.0. Keep that, so an upgrade does not cut off remote clients
        if [ -z "$(_read_env HOST)" ]; then
            _set_env HOST 0.0.0.0
            echo "Kept HOST=0.0.0.0 from the old default. Set HOST=127.0.0.1 in $ENV_FILE to allow only this machine."
        fi
    fi

    chmod 600 "$ENV_FILE"

    # Explicit options win over the config
    [ -n "$HOST_ARG" ] && _set_env HOST "$HOST_ARG"
    [ -n "$PORT_ARG" ] && _set_env PORT "$PORT_ARG"
    [ -n "$STORAGE_ARG" ] && _set_env STORAGE_PATH "$(realpath -m "$STORAGE_ARG")"

    if [ -z "$(_read_env PORT)" ]; then
        _set_env PORT 50051
    fi

    if [ -z "$(_read_env STORAGE_PATH)" ]; then
        _set_env STORAGE_PATH "$DEFAULT_STORAGE_DIR"
    fi
}

# Creates the keys file, or checks an existing one, and points the config at it
_write_keys() {
    local keys_file key

    keys_file="${KEYS_FILE_ARG:-$(_read_env KEYS_FILE)}"
    keys_file="$(realpath -m "${keys_file:-$DEFAULT_KEYS_FILE}")"

    if [ ! -e "$keys_file" ]; then
        mkdir -p "$(dirname "$keys_file")"

        # Configs from older installs hold the keys, move them instead of making new ones:
        # files encrypted with the old key would become unreadable
        local api_key encryption_key
        api_key="$(_read_env ROOT_API_KEY)"
        encryption_key="$(_read_env ENCRYPTION_KEY)"

        if [ -n "$encryption_key" ]; then
            echo "Moving the keys from $ENV_FILE to $keys_file..."
        else
            echo "Generating keys at $keys_file..."
            encryption_key="$(openssl rand -hex 32)"
            echo "IMPORTANT: Generated a random ENCRYPTION_KEY. Back it up: stored files cannot be decrypted without it."
        fi

        if [ -z "$api_key" ]; then
            api_key="$(openssl rand -hex 32)"
            GENERATED_API_KEY=1
        fi

        # umask in a subshell, so the file is created with mode 600 and is never readable by others
        (
            umask 077
            echo "ENCRYPTION_KEY=$encryption_key" > "$keys_file"
            echo "ROOT_API_KEY=$api_key" >> "$keys_file"
        )
    else
        echo "Using keys from $keys_file."
    fi

    [ -r "$keys_file" ] && [ -f "$keys_file" ] || _fail "$keys_file is not a readable file"

    # The server refuses a keys file that other users can read
    if [ "$(( 8#$(stat -c '%a' "$keys_file") & 8#077 ))" -ne 0 ]; then
        _fail "$keys_file is readable by other users, run: chmod 600 $keys_file"
    fi

    [ -n "$(_read_env ROOT_API_KEY "$keys_file")" ] || _fail "ROOT_API_KEY is missing in $keys_file"

    key="$(_read_env ENCRYPTION_KEY "$keys_file")"

    # Never replace an existing key: files encrypted with it would become unreadable
    if ! [[ "$key" =~ ^[0-9a-fA-F]{64}$ ]]; then
        _fail "ENCRYPTION_KEY in $keys_file must be 64 hex chars (32 bytes). Fix it manually. Generate a new key with: openssl rand -hex 32"
    fi

    local old_key
    old_key="$(_read_env ENCRYPTION_KEY)"

    # Dropping a different old key from the config would lose it for good
    if [ -n "$old_key" ] && [ "$old_key" != "$key" ]; then
        _fail "$ENV_FILE and $keys_file hold different ENCRYPTION_KEY values. Keep the one your files were encrypted with, and remove the other"
    fi

    # The server rejects a key set in both files
    sed -i '/^ROOT_API_KEY=/d; /^ENCRYPTION_KEY=/d' "$ENV_FILE"
    _set_env KEYS_FILE "$keys_file"
}

install() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --host) HOST_ARG="$2"; shift 2 ;;
            --port) PORT_ARG="$2"; shift 2 ;;
            --storage) STORAGE_ARG="$2"; shift 2 ;;
            --keys-file) KEYS_FILE_ARG="$2"; shift 2 ;;
            -h | --help) usage; exit 0 ;;
            *) usage; _fail "unknown option $1" ;;
        esac
    done

    if [ -n "$PORT_ARG" ] && ! { [[ "$PORT_ARG" =~ ^[0-9]+$ ]] && [ "$PORT_ARG" -ge 1 ] && [ "$PORT_ARG" -le 65535 ]; }; then
        _fail "--port must be a number between 1 and 65535"
    fi

    command -v openssl >/dev/null || _fail "openssl is required"

    echo "--- Installing Depozit ---"

    # Check systemd user session is available
    if ! systemctl --user status >/dev/null 2>&1; then
        echo "Error: systemd user session is not available." >&2
        echo "Make sure you are logged in as a regular user with a running session." >&2
        exit 1
    fi

    SRC_DIR="$(mktemp -d)"
    trap 'rm -rf "$SRC_DIR"' EXIT
    _fetch "$SRC_DIR"

    echo "Creating directories..."
    mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$SYSTEMD_DIR"
    chmod 700 "$CONFIG_DIR"

    _write_config
    _write_keys

    local storage_dir
    storage_dir="$(_read_env STORAGE_PATH)"
    mkdir -p "$storage_dir"

    echo "Installing binaries to $BIN_DIR..."

    # Stop first, so the running binary is not replaced under it
    systemctl --user stop "$SERVICE_NAME" 2>/dev/null || true

    command install -m 755 "$SRC_DIR/$SERVER_BINARY" "$BIN_DIR/$SERVER_BINARY"
    command install -m 755 "$SRC_DIR/$CLI_BINARY" "$BIN_DIR/$CLI_BINARY"

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
UMask=0077

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable "$SERVICE_NAME"
    systemctl --user start "$SERVICE_NAME"

    # Add ~/.local/bin to PATH in shell rc files if not already present
    _ensure_path

    local tls_cert host
    tls_cert="$(_read_env TLS_CERT)"
    host="$(_read_env HOST)"

    echo ""
    echo "--- Installation Complete ---"

    local keys_file
    keys_file="$(_read_env KEYS_FILE)"

    # Print the key only to a terminal, never into CI or deploy logs
    if [ -n "$GENERATED_API_KEY" ] && [ -t 1 ]; then
        echo "ROOT_API_KEY: $(_read_env ROOT_API_KEY "$keys_file")"
    else
        echo "ROOT_API_KEY: stored in $keys_file"
    fi

    echo "Host:         $host"
    echo "Port:         $(_read_env PORT)"
    echo "Storage:      $storage_dir"
    echo "Config:       $ENV_FILE"
    echo "Keys:         $keys_file"
    echo "TLS:          ${tls_cert:-disabled}"
    echo ""

    # TLS is set up by hand in the config, see "TLS" in the README
    if [ -z "$tls_cert" ] && [ "$host" != "127.0.0.1" ] && [ "$host" != "localhost" ] && [ "$host" != "::1" ]; then
        echo "WARNING: HOST=$host accepts remote clients without TLS. The API key and files travel in plaintext."
        echo "Set TLS_CERT and TLS_KEY in $ENV_FILE, then run: depozit stop && depozit start"
        echo ""
    fi

    echo "ENCRYPTION_KEY is stored in $keys_file. Back it up: stored files cannot be"
    echo "decrypted without it."
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

# Last line on purpose: a download cut short by curl runs nothing
install "$@"
