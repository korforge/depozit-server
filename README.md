# Depozit Server

A secure gRPC file storage server.

## Install

One command, no root required. It downloads the prebuilt server from `dist/` and the `depozit` management script from `scripts/`, both from the branch the script comes from, checks their sha256 sums, installs both to `~/.local/bin`, creates a config at `~/.depozit/.env`, generates the keys in `~/.depozit/keys.env`, and starts the service via systemd:


```bash
mkdir -p ~/.depozit
(umask 077; {
  echo "ENCRYPTION_KEY=$(openssl rand -hex 32)"
  echo "ROOT_API_KEY=$(uuidgen)"
} > ~/.depozit/keys.env)

curl -fsSL https://raw.githubusercontent.com/korforge/depozit-server/main/install.sh \
  | bash -s -- --port 50051 --keys-file ~/.depozit/keys.env
```

| Option | Description |
|--------|-------------|
| `--host HOST` | Bind address (default `127.0.0.1`: only this machine can connect) |
| `--port PORT` | gRPC port (default `50051`) |
| `--storage DIR` | Directory where files are stored (default `~/.depozit/storage`) |
| `--keys-file FILE` | Keys file (default `~/.depozit/keys.env`). Generated when it does not exist, used as is when it does |

The script URL picks the branch: `.../feat/streaming/install.sh` installs the binaries of `feat/streaming`. A piped script cannot see its own URL, so the pre-push hook writes the pushed branch into `BRANCH` in `install.sh` (see [Development](#development)).

To install a specific version, download that tag from GitHub and run the installer inside it. When `install.sh` finds `dist/` next to it, it installs those files instead of downloading:

```bash
curl -fsSL https://github.com/korforge/depozit-server/archive/refs/tags/v1.0.0.tar.gz | tar xz
bash depozit-server-1.0.0/install.sh --port 50051
```

For deploys, prefer a tag: `main` moves, so the same curl command can install different code later.

Running the installer again upgrades the binaries and keeps the existing config and keys. Configs from older installs that hold the keys themselves are migrated: the keys move to the keys file, unchanged. Options passed on the command line replace the matching config values.

To install from a clone instead, without downloading:

```bash
make install
```

## Development

`make build` builds the server for linux/amd64 and linux/arm64 into `dist/`, and writes `dist/sha256sums.txt` with the sums of both binaries and of `scripts/depozit`. `dist/` is committed. `install.sh` installs the server from `dist/` and the management script straight from `scripts/`. The build is reproducible: the same source and Go version give the same bytes, so `dist/` only changes when the code does.

Enable the git hooks once per clone:

```bash
make hooks
```

The `pre-push` hook runs `make build` before every push, and sets `BRANCH` in `install.sh` to the branch being pushed, so the curl command for each branch installs that branch's binaries. When either changed, it commits them as `build: update dist binaries for <branch>` and stops the push, because git does not include commits made during a push. Run `git push` again to push them. After merging a branch into `main`, the next push of `main` sets `BRANCH` back to `main` the same way. The hook refuses to run while `src/`, `scripts/`, `install.sh`, `go.mod` or `go.sum` have uncommitted changes, so the pushed files always match the pushed code.

To release, push first so the build commit lands, then tag that commit and push the tag:

```bash
git push && git push   # the second push sends the build commit, if the hook made one
git tag v1.0.0
git push origin v1.0.0
```

### Network and TLS

> **Disclaimer:** the installer does not set up TLS. By default the server listens on `127.0.0.1`, so only programs on the same machine can connect, and plaintext is safe. To accept clients from other machines, you must set up TLS yourself. Without it, the API key and all file contents travel in plaintext and anyone on the network path can read them.

To accept remote clients:

1. Get a certificate and its private key in PEM format, for example from Let's Encrypt or your internal CA.
2. Make sure your user can read both files, and that nobody else can read the key: `chmod 600 server.key`.
3. Edit `~/.depozit/.env`:

   ```bash
   HOST=0.0.0.0
   TLS_CERT=/absolute/path/server.crt
   TLS_KEY=/absolute/path/server.key
   ```

4. Restart: `depozit stop && depozit start`. The server refuses to start when the certificate or the key cannot be loaded. Check `journalctl --user -u depozit` if it does not come up.

A renewed certificate at the same path is used after a restart. Running behind a TLS-terminating proxy (nginx, Caddy, a load balancer) also works: keep `HOST=127.0.0.1` and let the proxy handle TLS.

Installs made before `127.0.0.1` became the default keep `HOST=0.0.0.0` when upgraded, so remote clients are not cut off. The installer warns when the server accepts remote clients without TLS.

### Keys

The secrets live in their own file, apart from the config. The config points at it with `KEYS_FILE`:

```bash
# ~/.depozit/keys.env, mode 600
ENCRYPTION_KEY=<64 hex chars>
ROOT_API_KEY=<random>
```

The server refuses to start when the keys file is readable by other users, and when a key is set in both the keys file and the config.

Never pass `ROOT_API_KEY` or `ENCRYPTION_KEY` as command-line arguments or `Environment=` lines: they show up in `ps`, shell history, CI logs and `systemctl show`. Instead:

- **Single host:** let the installer generate the keys file. The keys never leave the machine. The API key is printed only when the output is a terminal, never into CI logs.
- **Your own path** (an encrypted volume, a directory your backup covers, a file your secret manager writes): create the keys file there before installing, then pass it with `--keys-file`:

  ```bash
  (umask 077 && {
      echo "ENCRYPTION_KEY=$(openssl rand -hex 32)"
      echo "ROOT_API_KEY=$(openssl rand -hex 32)"
  } > /secure/depozit-keys.env)

  curl -fsSL https://raw.githubusercontent.com/korforge/depozit-server/main/install.sh \
    | bash -s -- --keys-file /secure/depozit-keys.env
  ```

  `umask 077` creates the file with mode `600` from the start, so the keys are never readable by other users, not even for a moment. A path that does not exist yet also works: the installer generates the file there the same way.

`depozit uninstall` deletes a keys file inside `~/.depozit` and keeps one stored anywhere else.

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
| `HOST`          | Bind address (default `127.0.0.1`). Use `0.0.0.0` to accept remote clients, with TLS |
| `PORT`          | gRPC port (required; installer sets `50051`) |
| `ROOT_API_KEY`  | API key generated on first install. Set it in the keys file |
| `ENCRYPTION_KEY`| 64 hex chars key used to encrypt stored files. Set it in the keys file and back it up |
| `KEYS_FILE` | Path to the keys file with `ROOT_API_KEY` and `ENCRYPTION_KEY`. Must not be readable by other users |
| `STORAGE_PATH`  | Directory where files are stored   |
| `TLS_CERT`, `TLS_KEY` | Paths to a PEM certificate and its key. Set both to serve over TLS, or neither to run without TLS. Without them the API key and files travel in plaintext, and the server logs a warning on a non-loopback `HOST` |
| `MAX_FILE_SIZE` | Largest upload in bytes (default `0`, no limit) |
| `MAX_CONCURRENT_STREAMS` | Concurrent RPCs per connection (default `100`) |

Each container is a directory inside `STORAGE_PATH`. The `default` container is created on startup and is used when a request does not set `container`.

The API key gives full access to every container, including creating, renaming and deleting them.

Files are encrypted with AES-256-GCM in 64 KiB chunks. Each chunk is bound to its file, its path in the container, and its position, so a truncated, reordered or edited file fails to read with `DATA_LOSS`. Renaming a file encrypts it again under the new path. Uploads in progress live in `<container>/.depozit-tmp/`, which is removed on startup and never listed.

## Uninstall

```bash
depozit uninstall
```

The uninstall command removes the service, binaries, and config. **Stored files are not deleted.** The command will print the storage path so you can remove it manually if needed.
