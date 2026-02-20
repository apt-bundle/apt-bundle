---
layout: default
title: Contributing
nav_order: 5
---

# Contributing

This guide is for developers who want to contribute to apt-bundle or understand its internals.

## Project Structure

```
apt-bundle/
├── cmd/
│   └── apt-bundle/       # Main entry point
├── internal/
│   ├── aptfile/          # Aptfile parsing
│   ├── commands/         # CLI commands (install, dump, check)
│   ├── apt/              # APT interactions
│   └── config/           # Configuration
├── docs/                 # Documentation
├── Makefile              # Build automation
└── go.mod                # Go module definition
```

## Prerequisites

- Go 1.21 or later
- Debian/Ubuntu-based system (for testing)
- `make` (for build automation)

## Building

### Build the Binary

```bash
make build
```

The binary will be created at `build/apt-bundle`.

### Install Locally

```bash
sudo make install
```

This installs to `/usr/local/bin/apt-bundle` by default. You can customize:

```bash
INSTALL_DIR=$HOME/.local/bin USE_SUDO="" make install
```

### Build Options

The Makefile uses the following build flags:
- `CGO_ENABLED=0`: Pure Go compilation (no C dependencies)
- `-ldflags="-s -w"`: Strip debug symbols for smaller binary size

## Development Workflow

### Format Code

```bash
make fmt
```

This runs `go fmt ./...` to format all Go code.

### Run Static Analysis

```bash
make vet
```

This runs `go vet ./...` to catch common errors.

### Run Linter

```bash
make lint
```

This runs `golangci-lint` for comprehensive code analysis.

### Run Tests

```bash
make test
```

This runs all tests with verbose output.

### Clean Build Artifacts

```bash
make clean
```

Removes the `build/` directory.

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/aptfile/...

# Run tests with coverage
go test -cover ./...
```

### Writing Tests

- Place test files alongside source files with `_test.go` suffix
- Use standard Go testing patterns
- Mock external commands (apt-get, add-apt-repository) for unit tests
- Use integration tests for end-to-end scenarios

### Test Requirements

- Tests should be idempotent when possible
- Avoid requiring root privileges for unit tests
- Use temporary directories for file operations
- Clean up after tests

## Testing GitHub Actions Locally

There are two ways to test the CI workflow locally before pushing:

### Option 1: Quick Test with Makefile (Recommended)

The easiest way is to use the `ci-build` Makefile target, which mimics the exact CI build step:

```bash
# Test with default architecture (amd64) and version from VERSION file
make ci-build

# Test with specific architecture
make ci-build ARCH=arm64

# Test with specific architecture and version
make ci-build ARCH=amd64 VERSION=0.1.5
```

This will:
- Build the binary for the specified architecture
- Create a .deb package using nfpm (with the same environment variables as CI)
- Rename the package to include architecture in filename
- Copy it to `artifacts/` directory

**Note:** This only tests the build step. It doesn't test the full workflow (version calculation, release creation, etc.).

### Option 2: Full Workflow Testing with `act`

For testing the complete GitHub Actions workflow, you can use [act](https://github.com/nektos/act):

**Installation:**

```bash
# On macOS
brew install act

# On Linux (using the install script)
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Or download from releases: https://github.com/nektos/act/releases
```

**Basic Usage:**

```bash
# List all workflows
act -l

# Run the build job (dry-run, won't actually build)
act -j build --dryrun

# Run the build job for a specific architecture
act -j build --matrix goarch:amd64 --matrix debarch:amd64

# Run with environment variables
act -j build -e VERSION=0.1.0
```

**Limitations:**

- `act` runs workflows in Docker containers, so it's slower than the Makefile approach
- Some GitHub Actions features may not work perfectly (secrets, API calls, etc.)
- The release job requires GitHub API access which may not work locally

**Recommended Workflow:**

1. **Quick iteration**: Use `make ci-build` to test build changes rapidly
2. **Full workflow**: Use `act` before pushing to test the complete workflow
3. **Final verification**: Push to a test branch and let GitHub Actions run

### Debugging CI Test Failures

When tests pass locally but fail in CI:

1. **Run tests in a CI-like environment** (Docker, matches ubuntu-latest):
   ```bash
   make test-docker
   ```
   This often reproduces the failure locally.

2. **Download the test output artifact** when CI fails: In the Actions run, open the "Run tests" step. If the log is truncated, go to the "Summary" tab and download the `test-output` artifact (uploaded on failure).

3. **Use `-failfast`** (already in CI): Tests stop at the first failure, so the failing test name and error appear at the end of the log.

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting (enforced by `make fmt`)
- Follow the guidelines enforced by `golangci-lint`
- Write clear, self-documenting code
- Add comments for exported functions and types

## Architecture

### Command Structure

The CLI uses [Cobra](https://github.com/spf13/cobra) for command parsing:

- `cmd/apt-bundle/main.go`: Entry point
- `internal/commands/`: Command implementations
  - `install.go`: Install command
  - `check.go`: Check command
  - `dump.go`: Dump command

### Aptfile Parsing

- `internal/aptfile/`: Parses Aptfile format
  - Tokenizes directives
  - Validates syntax
  - Provides structured representation

### APT Integration

- `internal/apt/`: Wraps apt-get and add-apt-repository
  - Executes system commands
  - Checks for existing packages/repositories
  - Handles errors gracefully

### Configuration

- `internal/config/`: Manages configuration
  - Default file paths
  - Command-line flags
  - Environment variables

## Pull Request Guidelines

### Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/apt-bundle.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `make test`
6. Format code: `make fmt`
7. Run linter: `make lint`
8. Commit your changes
9. Push to your fork
10. Open a Pull Request

### PR Guidelines

- Keep PRs focused on a single feature or fix
- Include tests for new functionality
- Update documentation as needed
- Ensure all tests pass
- Follow the existing code style

### Commit Messages

Write clear, descriptive commit messages:

```
Add support for version pinning in Aptfile

- Parse version strings from apt directives
- Pass version to apt-get install
- Add tests for version parsing
```

## Dependencies

### External Dependencies

- [spf13/cobra](https://github.com/spf13/cobra): CLI framework

### Development Dependencies

- [golangci/golangci-lint](https://github.com/golangci/golangci-lint): Linter

### Managing Dependencies

```bash
# Download dependencies
make deps

# Add a new dependency
go get <package>

# Update dependencies
go get -u ./...
go mod tidy
```

## Binary Characteristics

The apt-bundle binary is designed to be:

- **Self-contained**: Statically linked, no external `.so` or `.dll` files
- **Portable**: Works on any Linux system with the same architecture
- **Small**: Stripped debug symbols for minimal size
- **Fast**: Minimal overhead over native apt commands

## Release Process

1. Update version in code (if versioning is implemented)
2. Update CHANGELOG.md
3. Create a git tag
4. Build release binaries
5. Create GitHub release
6. Upload binaries

## Troubleshooting

### Build Issues

- Ensure Go 1.21+ is installed: `go version`
- Check `go.mod` is up to date: `go mod tidy`
- Clean and rebuild: `make clean && make build`

### Test Issues

- Some tests may require a Debian/Ubuntu system
- Integration tests may need root privileges (use sudo)
- Check that required system commands are available

### Development Environment

- Use `direnv` or similar for environment management
- Consider using Docker for isolated testing
- Use `go run` for quick iteration: `go run ./cmd/apt-bundle`

### CI and Workflow Issues

**nfpm not found:** The Makefile will auto-install nfpm if it's missing. For `act`, ensure Docker is running.

**Architecture-specific issues:** Test each architecture individually:

```bash
make ci-build ARCH=amd64
make ci-build ARCH=arm64
make ci-build ARCH=armhf
make ci-build ARCH=i386
```

**Version format:** The CI workflow expects versions in `MAJOR.MINOR.PATCH` format. The Makefile defaults to `VERSION.0` if only `MAJOR.MINOR` is in the VERSION file.

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://github.com/spf13/cobra/blob/main/site/content/introduction.md)
- [Go Testing](https://golang.org/pkg/testing/)
- [Project Requirements](https://github.com/apt-bundle/apt-bundle/blob/main/specs/requirements.md) - Internal specification
- [Technical Specification](https://github.com/apt-bundle/apt-bundle/blob/main/specs/tech-specs.md) - Internal specification

## Getting Help

- Open an issue on GitHub for bugs or feature requests
- Check existing issues and discussions
- Review the [API Reference](api-reference.html) for code documentation
