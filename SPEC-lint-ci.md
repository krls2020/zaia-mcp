# Spec: Code quality tooling + GitHub Actions pro ZAIA-MCP

Implementace podle vzoru `zaia/` (commit f8d5455). Adaptuj pro zaia-mcp.

---

## Co implementovat

6 souborů:

1. `.golangci.yaml` — konfigurace linteru
2. `Makefile` — build/test/lint targety
3. `tools/install.sh` — instalace golangci-lint
4. `.github/workflows/main.yml` — CI pipeline
5. `.github/workflows/release.yml` — produkční release
6. `.github/workflows/pre-release.yml` — pre-release

---

## 1. `.golangci.yaml`

Zkopíruj z `../zaia/.golangci.yaml` beze změn. Stejných 63 linterů, stejné exclude-rules.

Jediny rozdíl — pokud zaia-mcp nepoužívá nějaký pattern (např. `WalkDir` nilerr), příslušné nolint komentáře v kódu nebudou potřeba. Ale config je univerzální.

---

## 2. `Makefile`

```makefile
.PHONY: help test test-race lint vet build all

test:
	go test -v ./... -count=1

test-race:
	go test -race ./... -count=1

lint:
	GOOS=darwin GOARCH=arm64 golangci-lint run ./...
	GOOS=linux GOARCH=amd64 golangci-lint run ./...
	GOOS=windows GOARCH=amd64 golangci-lint run ./...

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
```

### Klíčové rozdíly oproti zaia:
- Binary prefix: `zaia-mcp-` (ne `zaia-`)
- Build target: `./cmd/zaia-mcp/main.go` (ne `./cmd/zaia/main.go`)

---

## 3. `tools/install.sh`

Identický se `../zaia/tools/install.sh`:

```bash
#!/bin/bash
cd "$(dirname "$0")"
set -e
cd ..
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
[[ ! -d "${GOBIN}" ]] && mkdir -p "${GOBIN}"
export GOBIN=$PWD/bin
export PATH="${GOBIN}:${PATH}"
echo "GOBIN=${GOBIN}"
rm -rf tmp
GOBIN="$GOBIN" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
```

Nezapomenout: `chmod +x tools/install.sh`

---

## 4. `.github/workflows/main.yml`

```yaml
name: Main

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - '*'

jobs:
  build:
    name: build && tests for ${{ matrix.name }}
    runs-on: ubuntu-22.04
    env:
      CGO_ENABLED: '0'
    strategy:
      matrix:
        include:
          - name: linux amd64
            env:
              GOOS: linux
              GOARCH: amd64
            runTests: true
          - name: linux 386
            env:
              GOOS: linux
              GOARCH: 386
            runTests: false
          - name: darwin amd64
            env:
              GOOS: darwin
              GOARCH: amd64
            runTests: true
          - name: darwin arm64
            env:
              GOOS: darwin
              GOARCH: arm64
            runTests: false
          - name: windows amd64
            env:
              GOOS: windows
              GOARCH: amd64
            runTests: true

    steps:
    - name: check out code into the Go module directory
      uses: actions/checkout@v4

    - name: set up Go 1.x
      uses: actions/setup-go@v5
      with:
        go-version-file: "go.mod"
      id: go

    - name: get dependencies
      run: |
        export GOPATH=$HOME/go
        ./tools/install.sh
        echo "$PWD/bin" >> $GITHUB_PATH

    - name: build
      env: ${{ matrix.env }}
      run: go build -v ./cmd/... ./internal/...

    - name: test
      if: ${{ matrix.runTests }}
      run: go test -v ./... -count=1

    - name: lint
      run: golangci-lint run ./...
```

### Kritické: `echo "$PWD/bin" >> $GITHUB_PATH`
Bez tohoto řádku golangci-lint nebude v PATH pro lint step. Toto byla chyba v první iteraci zaia CI.

---

## 5. `.github/workflows/release.yml`

```yaml
name: release

on:
  release:
    types:
      - released

jobs:
  build:
    name: build & upload ${{ matrix.name }}
    runs-on: ubuntu-22.04
    env:
      CGO_ENABLED: '0'
    strategy:
      matrix:
        include:
          - name: linux amd64
            env: { GOOS: linux, GOARCH: amd64 }
            file: zaia-mcp-linux-amd64
            compress: true
            strip: true
          - name: linux 386
            env: { GOOS: linux, GOARCH: 386 }
            file: zaia-mcp-linux-i386
            compress: true
            strip: true
          - name: darwin amd64
            env: { GOOS: darwin, GOARCH: amd64 }
            file: zaia-mcp-darwin-amd64
            compress: false
            strip: false
          - name: darwin arm64
            env: { GOOS: darwin, GOARCH: arm64 }
            file: zaia-mcp-darwin-arm64
            compress: false
            strip: false
          - name: windows amd64
            env: { GOOS: windows, GOARCH: amd64 }
            file: zaia-mcp-win-x64.exe
            compress: false
            strip: false

    steps:
      - name: checkout code
        uses: actions/checkout@v4

      - name: set up Go 1.x
        uses: actions/setup-go@v5
        with:
          go-version-file: "go.mod"
        id: go

      - name: get dependencies
        run: |
          export GOPATH=$HOME/go
          ./tools/install.sh
          echo "$PWD/bin" >> $GITHUB_PATH

      - name: build
        env: ${{ matrix.env }}
        run: >-
          go build
          -o builds/${{ matrix.file }}
          -ldflags "-s -w -X github.com/zeropsio/zaia-mcp/internal/server.Version=${{ github.event.release.tag_name }}"
          ./cmd/zaia-mcp/main.go

      - name: compress binary
        if: ${{ matrix.compress }}
        uses: svenstaro/upx-action@v2
        with:
          file: ./builds/${{ matrix.file }}
          strip: ${{ matrix.strip }}

      - name: upload asset clean bin
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ github.event.release.upload_url }}
          asset_path: ./builds/${{ matrix.file }}
          asset_name: ${{ matrix.file }}
          asset_content_type: application/octet-stream
```

### Version ldflags — POZOR

Aktuální verze v zaia-mcp je hardcoded v `internal/server/server.go`:
```go
srv := mcp.NewServer(&mcp.Implementation{
    Name:    "zaia-mcp",
    Version: "0.1.0",  // <-- toto
}, ...)
```

Pro ldflags je potřeba:
1. Extrahovat verzi do package-level proměnné:
   ```go
   // Version is set at build time via -ldflags
   var Version = "dev"
   ```
2. Použít ji v `NewWithExecutor()`:
   ```go
   Version: Version,
   ```
3. ldflags path: `-X github.com/zeropsio/zaia-mcp/internal/server.Version=$TAG`

---

## 6. `.github/workflows/pre-release.yml`

Stejné jako release.yml, ale:
- Trigger: `prereleased` (ne `released`)
- Build flags: `-tags devel` (ne `-s -w`)
- ldflags: `-X github.com/zeropsio/zaia-mcp/internal/server.Version=$TAG` (bez `-s -w`)
- Bez UPX komprese
- Jen upload čistých binárky

---

## Lint opravy

Po vytvoření souborů spusť `golangci-lint run ./...` lokálně a oprav všechny findings.

Typické vzory k opravě (ze zkušenosti s zaia):
- `errcheck` — `_ = json.Unmarshal(...)` v testech, `_ = writeJSON(...)` v produkci
- `usetesting` — `context.Background()` → `t.Context()` v testech (Go 1.24+)
- `errorlint` — `err.(*Type)` → `errors.As(err, &target)`
- `forcetypeassert` — vyloučeno v testech přes config
- `goconst` — opakované stringy → konstanty
- `gofmt` — `gofmt -w ./internal/ ./cmd/ ./integration/`

---

## Ověření

1. `golangci-lint run ./...` → 0 issues
2. `go test ./... -count=1` → all pass
3. `go build -o builds/zaia-mcp-test -ldflags "-s -w -X github.com/zeropsio/zaia-mcp/internal/server.Version=v0.0.1-test" ./cmd/zaia-mcp/main.go` → OK
4. Push → CI workflow zelený

---

## Referenční implementace

Viz `../zaia/` soubory (commit f8d5455):
- `.golangci.yaml` — plná konfigurace s exclude-rules
- `Makefile` — build targety
- `tools/install.sh` — instalátor
- `.github/workflows/main.yml` — CI
- `.github/workflows/release.yml` — release
- `.github/workflows/pre-release.yml` — pre-release

Git remote: `git@github.com:krls2020/zaia-mcp.git` (SSH)
