GOBIN := $(shell go env GOPATH)/bin
PATH := $(PATH):$(GOBIN)

# Binaries committed to the repo, downloaded by install.sh
DIST := dist

# Same source and Go version give the same bytes, so an unchanged build leaves dist/ unchanged.
# -buildvcs=false matters: the embedded commit hash would change the binaries on every commit
GO_BUILD := CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid="

.PHONY: protobufs build hooks install install-dev start test coverage

protobufs:
	@protoc --go_out=./src --go-grpc_out=./src src/protobufs/files.proto

build:
	@mkdir -p $(DIST)
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DIST)/depozit-server-linux-amd64 ./src
	@GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(DIST)/depozit-server-linux-arm64 ./src
	@# Paths relative to the repo root, the same paths install.sh downloads. The depozit script is
	@# installed straight from scripts/, its sum lives here so one file covers everything installed
	@sha256sum $(DIST)/depozit-server-linux-amd64 $(DIST)/depozit-server-linux-arm64 scripts/depozit > $(DIST)/sha256sums.txt

# Git does not version hooks, point it at the ones in the repo
hooks:
	@git config core.hooksPath .githooks
	@echo "Git hooks enabled from .githooks/"

install:
	@if [ -f "$(HOME)/.local/bin/depozit-server" ]; then \
		echo "Depozit is already installed. Run 'depozit uninstall' first."; \
	else \
		$(MAKE) --no-print-directory build && bash install.sh; \
	fi

install-dev:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@sudo apt install -y protobuf-compiler

# Local config for make start: files stored in data/, keys in their own file like an install
DEV_ENV := test.env
DEV_KEYS := data/keys.env

start: $(DEV_ENV) $(DEV_KEYS)
	@go run ./src -env $(DEV_ENV)

$(DEV_ENV):
	@printf "PORT=5051\nSTORAGE_PATH=data\nKEYS_FILE=$(DEV_KEYS)\n" > $@

# Fixed API key so local clients can hardcode it, the encryption key is random per clone
$(DEV_KEYS):
	@mkdir -p $(dir $@)
	@(umask 077; { echo "ENCRYPTION_KEY=$$(openssl rand -hex 32)"; echo "ROOT_API_KEY=123456"; } > $@)

test:
	@go test -v ./...

coverage:
	@go test ./... -coverpkg=./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out
