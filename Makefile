.PHONY: help test test-race lint vet build all windows-amd linux-amd linux-386 darwin-amd darwin-arm

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILT   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -X github.com/zeropsio/zaia-mcp/internal/server.Version=$(VERSION)

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	go test -v ./... -count=1

test-race: ## Run tests with race detection
	go test -race ./... -count=1

lint: ## Run linter for all platforms
	GOOS=darwin GOARCH=arm64 golangci-lint run ./... --verbose
	GOOS=linux GOARCH=amd64 golangci-lint run ./... --verbose
	GOOS=windows GOARCH=amd64 golangci-lint run ./... --verbose

vet: ## Run go vet
	go vet ./...

build: ## Build binary to bin/
	go build -ldflags "$(LDFLAGS)" -o bin/zaia-mcp ./cmd/zaia-mcp

#########
# BUILD #
#########
all: windows-amd linux-amd linux-386 darwin-amd darwin-arm ## Cross-build all platforms

windows-amd: ## Build for Windows amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-mcp-win-x64.exe ./cmd/zaia-mcp/main.go

linux-amd: ## Build for Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-mcp-linux-amd64 ./cmd/zaia-mcp/main.go

linux-386: ## Build for Linux 386
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o builds/zaia-mcp-linux-i386 ./cmd/zaia-mcp/main.go

darwin-amd: ## Build for macOS amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-mcp-darwin-amd64 ./cmd/zaia-mcp/main.go

darwin-arm: ## Build for macOS arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-mcp-darwin-arm64 ./cmd/zaia-mcp/main.go
