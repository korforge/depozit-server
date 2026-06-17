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
depozit status     # show status, port, and storage path
depozit uninstall  # remove everything except stored files
```

## Config

```bash
cat ~/.depozit/.env
```

| Variable        | Description                        |
|-----------------|------------------------------------|
| `PORT`          | gRPC port (default `50051`)        |
| `ROOT_API_KEY`  | API key generated on first install |
| `STORAGE_PATH`  | Directory where files are stored   |
| `CONTAINER`     | Default container name             |

## Uninstall

```bash
depozit uninstall
```

The uninstall command removes the service, binaries, and config. **Stored files are not deleted.** The command will print the storage path so you can remove it manually if needed.
