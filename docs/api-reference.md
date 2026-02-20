---
layout: default
title: API Reference
nav_order: 6
---

# API Reference

This page provides links to the Go package documentation for apt-bundle.

## Package Documentation

The complete API documentation is automatically generated and hosted on [pkg.go.dev](https://pkg.go.dev/):

**[View apt-bundle packages on pkg.go.dev](https://pkg.go.dev/github.com/apt-bundle/apt-bundle)**

## Package Structure

### Main Package

- **`cmd/apt-bundle`**: Main entry point for the CLI application

### Internal Packages

- **`internal/aptfile`**: Aptfile parsing and validation
  - Tokenizes Aptfile directives
  - Validates syntax
  - Provides structured representation of Aptfile contents

- **`internal/commands`**: CLI command implementations
  - `install`: Install packages from Aptfile
  - `check`: Verify packages are installed
  - `dump`: Generate Aptfile from system state

- **`internal/apt`**: APT system integration
  - Wraps `apt-get` commands
  - Manages repositories via `add-apt-repository`
  - Handles GPG key management
  - Checks package/repository status

- **`internal/config`**: Configuration management
  - Default file paths
  - Command-line flag parsing
  - Environment variable handling

## Key Types and Functions

### Aptfile Parsing

The `internal/aptfile` package provides:

- **Parser**: Parses Aptfile format into structured directives
- **Directive Types**: `AptDirective`, `PPADirective`, `DebDirective`, `KeyDirective`
- **Validation**: Ensures correct syntax and format

### Command Execution

The `internal/commands` package implements:

- **Install Command**: Orchestrates package installation
- **Check Command**: Verifies system state against Aptfile
- **Dump Command**: Generates Aptfile from current system

### APT Integration

The `internal/apt` package provides:

- **Package Management**: Install, check, list packages
- **Repository Management**: Add PPAs and deb repositories
- **Key Management**: Download and install GPG keys
- **Idempotency Checks**: Verify if packages/repositories already exist

## CLI Framework

apt-bundle uses [Cobra](https://github.com/spf13/cobra) for CLI parsing:

- Root command: `apt-bundle`
- Subcommands: `install`, `check`, `dump`
- Global flags: `--file`, `--help`, `--version`

## Example Usage (Programmatic)

While apt-bundle is primarily a CLI tool, the internal packages can be used programmatically:

```go
import (
    "github.com/apt-bundle/apt-bundle/internal/aptfile"
    "github.com/apt-bundle/apt-bundle/internal/apt"
)

// Parse an Aptfile
parser := aptfile.NewParser()
directives, err := parser.ParseFile("Aptfile")

// Check if a package is installed
installed, err := apt.IsPackageInstalled("vim")

// Add a PPA
err := apt.AddPPA("ppa:ondrej/php")
```

**Note:** The internal packages are not part of the public API and may change without notice. For programmatic use, consider opening an issue to discuss a stable API.

## Documentation Standards

- All exported functions and types have Go doc comments
- Examples are provided where applicable
- Error handling is documented
- Package-level documentation explains purpose and usage

## Generating Documentation Locally

To generate documentation locally:

```bash
# Install godoc
go install golang.org/x/tools/cmd/godoc@latest

# Run local documentation server
godoc -http=:6060

# View at http://localhost:6060/pkg/github.com/apt-bundle/apt-bundle/
```

## Related Documentation

- [Contributing](contributing.html) - Development setup and workflow
- [Technical Specification](https://github.com/apt-bundle/apt-bundle/blob/main/specs/tech-specs.md) - Internal specification
- [Requirements](https://github.com/apt-bundle/apt-bundle/blob/main/specs/requirements.md) - Internal specification

