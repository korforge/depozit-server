# Depozit Server

A secure gRPC file storage server.

## Install

Build and install (no root required):

```bash
make install
```

This compiles the server, installs both binaries to `~/.local/bin`, creates a config at `~/.depozit/.env` with a generated API key, and starts the service via systemd.

> If `depozit` is not found after install, open a new terminal or run `source ~/.bashrc`.

## Usage

```bash
depozit start      # start the service
depozit stop       # stop the service
depozit status     # show status, port, storage path, and container count
depozit containers                 # list containers
depozit containers create <name>   # create a container
depozit containers delete <name>   # delete a container and its files (asks for confirmation)
depozit uninstall  # remove everything except stored files
```

## Config

```bash
cat ~/.depozit/.env
```

| Variable        | Description                        |
|-----------------|------------------------------------|
| `HOST`          | Bind address (default `0.0.0.0`)   |
| `PORT`          | gRPC port (required; installer sets `50051`) |
| `ROOT_API_KEY`  | API key generated on first install |
| `ENCRYPTION_KEY`| 64 hex chars key used to encrypt stored files. Back it up |
| `STORAGE_PATH`  | Directory where files are stored   |

Each container is a directory inside `STORAGE_PATH`. The `default` container is created on startup and is used when a request does not set `container`.

## Uninstall

```bash
depozit uninstall
```

The uninstall command removes the service, binaries, and config. **Stored files are not deleted.** The command will print the storage path so you can remove it manually if needed.
