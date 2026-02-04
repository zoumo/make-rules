# Make Rules

A Go CLI tool for managing Go builds, containers, go.mod operations, and versioning.

## Installation

Build from source:

```bash
go build ./cmd/make-rules
```

Or run directly:

```bash
go run ./cmd/make-rules/main.go [command]
```

## CLI Overview

```bash
make-rules go build          # Build Go binaries
make-rules go install        # Install Go binaries
make-rules go mod tidy       # Tidy go.mod
make-rules go mod require    # Add module dependency
make-rules go mod replace    # Replace module dependency
make-rules go mod update     # Update module dependencies
make-rules go format         # Format Go code
make-rules go unittest       # Run unit tests
make-rules container build   # Build Docker images
make-rules version           # Show version
```

## Configuration

A `make-rules.yaml` file in the working directory configures the CLI:

```yaml
version: 1
go:
  minimumVersion: 1.14
  build:
    platforms:
      - linux/amd64
      - darwin/arm64
    globalHooksDir: hack/hooks
    flags: ["-v"]
    ldflags: []
    gcflags: []
  mod:
    require:
      - path: github.com/example/foo
        version: v1.0.0
        skipDeps: false
    replace:
      - path: github.com/example/foo
        newPath: github.com/example/foo
        version: v1.0.0
  format:
    local: github.com/zoumo/make-rules
    exclude:
      dirs:
        - output
        - hack/
      files:
        - datafile.go
        - bindata.go
        - .*_skip_format
  test:
    exclude:
      - testdata
container:
  imagePrefix: "prefix_"
  imageSuffix: "_suffix"
  registries:
    - hub.docker.io/zoumo
```

## Commands

### Build

`make-rules go build [target...]`

Cross-compile Go binaries for multiple platforms. Targets are discovered from directories containing `main.go` under `cmd/`.

Flags:
- `--platforms`: Target platforms (default from config)
- `--version`: Override version tag

The binary is placed at `<workspace>/bin/<GOOS>_<GOARCH>/<target>`.

Example:
```bash
make-rules go build
make-rules go build cmd/server
make-rules go build --platforms=linux/amd64
```

### Install

`make-rules go install`

Install built Go binaries to `GOBIN` or `GOPATH/bin`.

### Module Operations

`make-rules go mod tidy` - Clean up go.mod and go.sum
`make-rules go mod require <module> <version>` - Add a module dependency
`make-rules go mod replace <module> [new-path] <version>` - Replace a module dependency
`make-rules go mod update` - Update module dependencies based on config

### Format

`make-rules go format`

Format Go source code using goimports. Reads `go.format` config for local module and exclusions.

### Test

`make-rules go unittest`

Run unit tests for all packages. Reads `go.test.exclude` config to exclude directories.

### Container

`make-rules container build [target...]`

Build Docker images from `build/*/Dockerfile` directories.

Flags:
- `--registries`: Docker registries to tag images with
- `--version`: Override version tag

Image tag format: `<image-prefix><target><image-suffix>:<version>`

### Version

`make-rules version [--json]`

Show CLI version information including git commit, tree state, and build metadata.

## Version Handling

Version is determined automatically from git:

- Tagged commit: `v0.0.3`
- Dirty tree: `v0.0.3-dirty`
- Commits after tag: `v0.0.3-1+a1b2c3d`

Version is injected via ldflags during build using the `version` package.

## Development

Build:
```bash
go build ./...
```

Test:
```bash
go test ./...
```

Lint:
```bash
golangci-lint run
```

## Go Version Check

All `go` commands enforce the minimum Go version specified in `go.minimumVersion` config before execution.