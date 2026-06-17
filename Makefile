GOBIN := $(shell go env GOPATH)/bin
PATH := $(PATH):$(GOBIN)

.PHONY: build start install

protobufs:
	@protoc --go_out=./src --go-grpc_out=./src src/protobufs/files.proto

compile:
	@go build -o depozit-server ./src

install: compile
	@bash depozit.sh

install-dev:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@sudo apt install -y protobuf-compiler

start:
	@go run ./src -env test.env

test:
	@go test -v ./tests/...

coverage:
	@go test ./tests/... -coverpkg=./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out