.PHONY: help test test-race lint vet build all

test:
	go test -v ./... -count=1

test-race:
	go test -race ./... -count=1

lint:
	GOOS=darwin GOARCH=arm64 golangci-lint run ./... --verbose
	GOOS=linux GOARCH=amd64 golangci-lint run ./... --verbose
	GOOS=windows GOARCH=amd64 golangci-lint run ./... --verbose

vet:
	go vet ./...

build:
	go build -o ./zaia-mcp ./cmd/zaia-mcp

#########
# BUILD #
#########
all: windows-amd linux-amd linux-386 darwin-amd darwin-arm

windows-amd:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o builds/zaia-mcp-win-x64.exe ./cmd/zaia-mcp/main.go

linux-amd:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o builds/zaia-mcp-linux-amd64 ./cmd/zaia-mcp/main.go

linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o builds/zaia-mcp-linux-i386 ./cmd/zaia-mcp/main.go

darwin-amd:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o builds/zaia-mcp-darwin-amd64 ./cmd/zaia-mcp/main.go

darwin-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o builds/zaia-mcp-darwin-arm64 ./cmd/zaia-mcp/main.go
