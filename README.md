# apt-bundle

A declarative, Brewfile-like wrapper for `apt`, inspired by `brew bundle` — not a full config management system.

**[📚 Full Documentation](https://apt-bundle.github.io/apt-bundle/)** | [Installation](#installation) | [Usage](#usage)

## Overview

`apt-bundle` provides a simple, declarative, and shareable way to manage apt packages and repositories on Debian-based systems. Define your system dependencies in an `Aptfile` and install them with a single command.

## Features

- 📦 **Declarative Package Management**: Define packages in a simple text file
- 🔄 **Idempotent Operations**: Safe to run multiple times
- 🔀 **Sync**: Make system match Aptfile in one command (install + cleanup)
- 🔑 **Repository & Key Management**: Add PPAs, custom repositories, and GPG keys
- 📝 **Version Pinning**: Install specific package versions
- 🚀 **Simple CLI**: Easy-to-use command-line interface

## Why apt-bundle?

**Why not just bash scripts?** Idempotency is hard to get right; repository and key management is error-prone; and scripts become unmaintainable as they grow. apt-bundle gives you a single, declarative file and predictable behavior every time.

**Comparison to alternatives:**

| vs | apt-bundle advantage |
|----|------------------------|
| `dpkg --get-selections` | Human-readable Aptfile format, handles repos and keys, supports partial adoption |
| Ansible / Chef | Zero learning curve, no YAML or DSL—just packages and directives |
| Nix | Works with your existing apt ecosystem; no paradigm shift |

**Key benefits:** The Aptfile is declarative and shareable (commit it to git). Use `apt-bundle dump` to generate an Aptfile from your current system, `apt-bundle check` to validate without installing, `apt-bundle sync` to make the system match the Aptfile in one command (install + cleanup), and `apt-bundle cleanup` to remove packages no longer in the Aptfile (when using state-tracked installs).

## Installation

### Quick Install (Recommended)

Install the latest release using the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | sudo bash
```

### Manual Installation from .deb Package

Download and install the appropriate `.deb` package for your architecture:

```bash
# Detect your architecture
ARCH=$(dpkg --print-architecture)

# Download latest release (replace v1.0.0 with actual version)
VERSION=$(curl -s https://api.github.com/repos/apt-bundle/apt-bundle/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -LO https://github.com/apt-bundle/apt-bundle/releases/download/${VERSION}/apt-bundle_${VERSION#v}_linux_${ARCH}.deb

# Install
sudo dpkg -i apt-bundle_${VERSION#v}_linux_${ARCH}.deb
sudo apt-get install -f  # Install dependencies if needed
```

### From Source

```bash
# Clone the repository
git clone https://github.com/apt-bundle/apt-bundle.git
cd apt-bundle

# Build and install
make build
sudo make install
```

The binary will be installed to `/usr/local/bin/apt-bundle`.

### Building

```bash
# Build the binary
make build

# The binary will be in build/apt-bundle
./build/apt-bundle --help
```

## Usage

### Basic Commands

```bash
# Install packages from Aptfile (default: ./Aptfile)
sudo apt-bundle

# or explicitly
sudo apt-bundle install

# Use a different Aptfile
sudo apt-bundle --file /path/to/Aptfile

# Skip updating package lists (useful in CI/CD)
sudo apt-bundle --no-update

# See what would be installed/added without making changes
sudo apt-bundle install --dry-run

# Make system match Aptfile (install missing, remove no-longer-listed)
sudo apt-bundle sync
# See what would be installed/removed without making changes
sudo apt-bundle sync --dry-run

# Check if packages/repos/keys from Aptfile are present (exit 0 only if all present)
apt-bundle check
# Machine-friendly output for CI
apt-bundle check --json

# Validate Aptfile and check environment (apt-get, add-apt-repository, state)
apt-bundle doctor
# Only validate Aptfile (no environment checks)
apt-bundle doctor --aptfile-only

# List packages with available upgrades (exit 1 if any; for CI)
apt-bundle outdated

# Generate an Aptfile from current system
apt-bundle dump > Aptfile
```

Note: `dump` emits repository lines but not key directives for repos that use Signed-By; you may need to add those manually when installing from a dumped Aptfile.

### Aptfile Format

The `Aptfile` is a simple line-oriented text file with the following directives:

#### Install Packages

```aptfile
# Install latest version
apt vim
apt curl
apt git

# Install specific version
apt "nano=2.9.3-2"
```

#### Add PPAs

```aptfile
ppa ppa:ondrej/php
apt php8.1
```

#### Add Custom Repositories

```aptfile
# Add GPG key
key https://download.docker.com/linux/ubuntu/gpg

# Add repository
deb "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"

# Install packages from the repository
apt docker-ce
apt docker-ce-cli
```

### Complete Example

```aptfile
# Core development tools
apt build-essential
apt curl
apt git
apt vim
apt htop

# Specific version
apt "nano=2.9.3-2"

# PHP from PPA
ppa ppa:ondrej/php
apt php8.1
apt php8.1-cli
apt php8.1-fpm

# Docker
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
apt docker-ce
apt docker-ce-cli
apt containerd.io
```

## Use Cases

### Developer Onboarding

```bash
# Clone project
git clone https://github.com/myorg/myproject.git
cd myproject

# Install all system dependencies
sudo apt-bundle
```

### Dockerfile

Use `apt-bundle` in your Dockerfiles to manage system dependencies declaratively. See the [examples directory](examples/) for complete working examples:

- **[1-via-install-sh](examples/1-via-install-sh/)** - Install apt-bundle using the install script
- **[2-via-ppa](examples/2-via-ppa/)** - Install apt-bundle from the PPA
- **[3-via-apt-get](examples/3-via-apt-get/)** - Install apt-bundle via apt-get from the APT repository
- **[4-complex-via-apt-get](examples/4-complex-via-apt-get/)** - Multi-stage build with separate build/runtime Aptfiles

### System Sync

```bash
# On primary workstation
apt-bundle dump > Aptfile

# On new laptop
sudo apt-bundle
```

## Development

### Project Structure

```
apt-bundle/
├── cmd/
│   └── apt-bundle/       # Main entry point
├── internal/
│   ├── apt/              # APT interactions (packages, repos, keys)
│   ├── aptfile/          # Aptfile parsing
│   └── commands/         # CLI commands (install, dump, check)
├── examples/             # Docker examples by installation method
├── docs/                 # Documentation site and APT repository
├── specs/                # Requirements and technical specifications
├── Makefile              # Build automation
└── go.mod                # Go module definition
```

### Building and Testing

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run tests
make test

# Build
make build

# Install locally for testing
sudo make install
```

### Pre-commit Hook

Install the pre-commit hook to automatically format, lint, and verify Go code before each commit:

```bash
make install-hooks
```

The hook runs `golangci-lint --fix`, verifies the build, and runs lint. It only runs when Go files are staged (docs-only commits are skipped).

### Requirements

- Go 1.21 or later
- Debian/Ubuntu-based system (for running the tool)

### Version Management

The project uses a VERSION file for version management:
- The `VERSION` file contains the major.minor version (e.g., `1.0`)
- Patch versions are automatically incremented on each release
- To update the major or minor version, edit the `VERSION` file
- Releases are automatically created when code is merged to the `main` branch

## Technical Details

### Binary Characteristics

- **Self-contained**: The Go binary is statically linked and doesn't require external `.so` or `.dll` files
- **CGO_ENABLED=0**: Ensures pure Go compilation without C dependencies
- **Small size**: Compiled with `-ldflags="-s -w"` to strip debug symbols

### Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework

## Documentation

📚 **[Full Documentation Site](https://apt-bundle.github.io/apt-bundle/)** - Complete user guide, developer documentation, and API reference

For internal specifications:
- [Requirements](specs/requirements.md) - Detailed functional requirements
- [Technical Specification](specs/tech-specs.md) - Aptfile format and implementation details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Core functionality (install, dump, check)
- [x] Package version handling
- [x] PPA management
- [x] Custom repository management
- [x] GPG key management
- [x] CI/CD pipeline
- [x] Release automation
- [ ] Expand test coverage

