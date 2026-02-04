# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`make-rules` is a Go CLI tool for managing Go builds (with cross-compilation), container builds, go.mod operations, and versioning. It reads configuration from `make-rules.yaml` in the working directory.

## Development Commands

```bash
# Build the CLI
go build ./cmd/make-rules

# Run the CLI
go run ./cmd/make-rules/main.go [command]

# Test
go test ./...

# Lint (uses golangci-lint v2)
golangci-lint run
```

## Architecture

### CLI Pattern

This codebase uses `github.com/zoumo/golib/cli` patterns:

- **CommonOptions** (`pkg/cli/common/options.go`): Base struct embedded in all commands. Provides Logger, Workspace, and Config. `Complete()` loads config from `make-rules.yaml`.
- **Command interfaces**: Commands implement `cli.Command` and `cli.ComplexOptions` interfaces with `BindFlags()`, `Complete()`, `Validate()`, and `Run()` methods.
- **cli.NewCobraCommand()**: Wrapper that wires up Cobra hooks to call the interface methods in proper order.

### Config Loading Two-Tier Pattern

Config is loaded at two levels:

1. **Group commands** (`go`, `container`): Load config in `PersistentPreRunE` for validation/logging (runs before subcommands)
2. **Subcommands** (`go build`, `container build`, etc.): Load config via `CommonOptions.Complete()` during `PreRunE`

This exists because `PersistentPreRunE` runs before subcommand initialization. Group commands load config for shared validation (e.g., Go version check), subcommands load it for their own use.

### Command Structure

```
cmd/make-rules/
├── main.go           # Entry point
└── app/
    ├── root.go       # Root command, global flags
    ├── go.go         # "go" group command (PersistentPreRunE: Go version check)
    └── container.go  # "container" group command

pkg/cli/cmd/
├── golang/           # Go subcommands (build, install, format, unittest, mod/*)
├── container/        # Container subcommands (build)
└── utils/            # Shared CLI utilities (FindTargetsFrom, FilterTargets)
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/config` | Load `make-rules.yaml` configuration file (`config.Load()`), defines Config struct |
| `pkg/golang` | Go module operations (mod tidy, require, replace), Go version checking (`VerifyGoVersion()`) |
| `pkg/git` | Git integration (describe, tree state, semantic version via `SemanticVersion()`) |
| `pkg/runner` | Execute external commands with environment control (`runner.Runner`) |
| `pkg/cli/common` | CommonOptions for all commands |
| `version` | Build version info injection via ldflags (base.go contains placeholders) |

## CLI Commands Reference

### Root
- `make-rules` - Main command
- Global flags: `--color`, `--v` (log verbosity)

### Go Commands
| Command | Purpose |
|---------|---------|
| `go build` | Build Go binaries for multiple platforms (`--platforms`, `--version`) |
| `go install` | Install built Go binaries to GOBIN/GOPATH/bin |
| `go mod tidy` | Clean up go.mod and go.sum |
| `go mod require <module> <version>` | Add module dependency |
| `go mod replace <module> [new-path] <version>` | Replace module dependency |
| `go mod update` | Update module dependencies from config |
| `go format` | Format Go code using goimports |
| `go unittest` | Run unit tests (excludes dirs from `go.test.exclude`) |

### Container Commands
| Command | Purpose |
|---------|---------|
| `container build` | Build Docker images from `build/*/Dockerfile` (`--registries`, `--version`) |

### Version
| Command | Purpose |
|---------|---------|
| `version [--json]` | Show version info (git commit, tree state, build date) |

## Configuration File Structure

`make-rules.yaml`:

```yaml
version: 1
go:
  minimumVersion: "1.14"          # Minimum Go version, checked by `go` commands
  build:
    platforms: ["linux/amd64"]   # Cross-compile targets
    globalHooksDir: "hack/hooks" # Pre/post build hooks
    flags: []                     # Extra go build flags
    ldflags: []                   # Go linker flags
    gcflags: []                   # Go compiler flags
  mod:
    require: []                   # Default module dependencies
    replace: []                   # Default module replacements
  format:
    local: "module/path"          # Local module name for goimports
    exclude:
      dirs: []
      files: []
  test:
    exclude: []                   # Directories to skip in tests
container:
  registries: []                  # Docker registries to push tags to
  imagePrefix: ""                 # Image name prefix
  imageSuffix: ""                 # Image name suffix
```

## Version Handling

Version is determined at runtime from git via `pkg/git/Describe().SemanticVersion()`:

- Tagged commit: `v0.0.3`
- Dirty tree: `v0.0.3-dirty`
- Commits after tag: `v0.0.3-1+a1b2c3d` (1 commit after tag, full commit hash)

Version is injected via ldflags during build into `version/base.go` placeholders.

## Build Output Locations

- Go binaries: `<workspace>/bin/<GOOS>_<GOARCH>/<target>`
- Docker images: `<image-prefix><target><image-suffix>:<version>`

## Notes

- Go 1.23.0 required (see go.mod)
- Only English is allowed in code
- Uses testify for testing
- Target discovery: `cmd/*/main.go` for go build, `build/*/Dockerfile` for container build