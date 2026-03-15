# Dockerfile Examples with Aptfile

This directory contains practical examples demonstrating different ways to use `apt-bundle` with Dockerfiles.

## Examples Overview

The examples are organized by **installation method**, **repository type**, and **complexity**:

### 1. Via install.sh Script (`1-via-install-sh/`)
**Simplest approach** - Quick installation using the install script.

- ✅ Fastest to get started
- ✅ One-liner installation
- ✅ Always gets latest version
- ⚠️ No version pinning

**Best for:** Quick prototyping, development environments, scripts

### 2. Via PPA (`2-via-ppa/`)
**PPA-focused example** - Demonstrates using Personal Package Archives (PPAs).

- ✅ Simple syntax - no GPG key management needed
- ✅ Automatic GPG key handling
- ✅ Common use case for Ubuntu packages
- ✅ Perfect for community-maintained packages

**Best for:** Installing packages from PPAs (PHP, Git, Node.js, etc.)

### 3. Via APT Repository with Custom Repo (`3-via-apt-get/`)
**Production-ready approach** - Install from official APT repository; demonstrates the
`key` + `deb` pattern for adding a custom third-party repository (Docker CLI).

- ✅ Version control
- ✅ Better caching
- ✅ Multi-architecture support
- ✅ Custom repositories with automatic GPG key management

**Best for:** Production Docker images, CI/CD pipelines, any package from a custom repo

### 4. Complex via APT Repository (`4-complex-via-apt-get/`)
**Advanced pattern** - Multi-stage builds with many dependencies using APT repository.

- ✅ Multi-stage build pattern
- ✅ Many dependencies (build + runtime)
- ✅ Optimized production images
- ✅ Clear separation of concerns

**Best for:** Production applications, complex builds, optimized deployments

### 5. With Lock File (`5-with-lockfile/`)
**Reproducible builds** - Demonstrates the `Aptfile.lock` workflow for pinning every
package to an exact version.

- ✅ Fully reproducible installs
- ✅ Version drift detection
- ✅ Explicit audit trail of installed versions
- ✅ Works with any installation method

**Best for:** CI/CD pipelines, team environments, production Docker images where you need
the exact same package versions on every build

### 6. GitHub CLI (`6-github-cli/`)
**Real-world third-party repo** - Installs GitHub CLI (`gh`) from its official APT
repository; demonstrates the `key` + `deb` pattern and the new `signed-by=` option
support for copy-paste compatibility with official installation docs.

- ✅ Custom repo GPG key (`key` directive)
- ✅ Custom `deb` repository
- ✅ `signed-by=` in deb options (copy-paste from official docs)
- ✅ Traditional equivalent for comparison

**Best for:** Any package from a vendor-hosted APT repository that ships official
installation instructions with `signed-by=` in the deb line

## Quick Start

### Example 1: Simple install.sh approach
```bash
cd 1-via-install-sh
make build
make run
```

### Example 2: PPA usage
```bash
cd 2-via-ppa
make build
make run
```

### Example 3: APT repository with custom repos
```bash
cd 3-via-apt-get
make build
make run
```

### Example 4: Complex multi-stage
```bash
cd 4-complex-via-apt-get
make build
make run
```

### Example 5: Reproducible builds with lock file
```bash
cd 5-with-lockfile
make build
make run
```

### Example 6: GitHub CLI
```bash
cd 6-github-cli
make build
make test
```

## Feature Coverage Matrix

| Feature | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Basic `apt` packages | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| PPA repositories (`ppa`) | | ✅ | | | | |
| Custom repo GPG key (`key`) | | | ✅ | | | ✅ |
| Custom `deb` repository | | | ✅ | | | ✅ |
| `signed-by=` in deb options | | | | | | ✅ |
| Multi-stage build | | | | ✅ | | |
| Separate Aptfiles per stage | | | | ✅ | | |
| Lock file (`Aptfile.lock`) | | | | | ✅ | |
| Reproducible `--locked` install | | | | | ✅ | |
| Version pinning in Aptfile | | | | | ✅ | |
| `install.sh` setup | ✅ | ✅ | | | | |
| APT repo setup | | | ✅ | ✅ | ✅ | ✅ |

## Repository Types Comparison

| Type | Syntax | GPG Key | Use Case |
|------|--------|---------|----------|
| **PPA** | `ppa ppa:user/repo` | ✅ Automatic | Ubuntu community packages |
| **Custom Repo** | `key url` + `deb "url"` | ✅ Automatic | Official vendor repos (Docker, Node.js, etc.) |

## Installation Methods Comparison

| Method | Speed | Version Control | Production Ready | Multi-Arch |
|--------|-------|----------------|------------------|------------|
| install.sh | ⚡ Fastest | ❌ No | ⚠️ Limited | ✅ Yes |
| APT Repository | ⚡ Fast | ✅ Yes | ✅ Yes | ✅ Yes |

## Common Patterns

All examples follow these patterns:

1. **Install apt-bundle**: Either via install.sh or APT repository
2. **Copy Aptfile**: Place the Aptfile(s) in the container
3. **Run apt-bundle**: Install dependencies from the Aptfile
4. **Clean up**: Remove apt cache to reduce image size (where applicable)

## Notes

- All examples use `ubuntu:22.04` as the base image, but can be adapted for other Debian-based distributions
- The `DEBIAN_FRONTEND=noninteractive` environment variable prevents interactive prompts during package installation
- The APT repository examples use `[trusted=yes]` because the repository is currently unsigned
- PPAs are Ubuntu-specific and won't work on pure Debian systems
- Each example can be customized to fit your specific needs
