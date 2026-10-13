GOBIN := $(shell go env GOPATH)/bin
PATH := $(PATH):$(GOBIN)

.PHONY: protobufs compile install install-dev start test coverage

protobufs:
	@protoc --go_out=./src --go-grpc_out=./src src/protobufs/files.proto

compile:
	@go build -o depozit-server ./src

install:
	@if [ -f "$(HOME)/.local/bin/depozit-server" ]; then \
		echo "Depozit is already installed. Run 'depozit uninstall' first."; \
	else \
		$(MAKE) --no-print-directory compile && bash install.sh; \
	fi

install-dev:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@sudo apt install -y protobuf-compiler

start:
	@mkdir -p data
	@go run ./src -env test.env

test:
	@go test -v ./...

coverage:
	@go test ./... -coverpkg=./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out
