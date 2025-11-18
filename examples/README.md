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

### 3. Via APT Repository (`3-via-apt-get/`)
**Production-ready approach** - Install from official APT repository with custom repos.

- ✅ Version control
- ✅ Better caching
- ✅ Multi-architecture support
- ✅ Custom repositories with GPG keys

**Best for:** Production Docker images, CI/CD pipelines, custom repositories

### 4. Complex via APT Repository (`4-complex-via-apt-get/`)
**Advanced pattern** - Multi-stage builds with many dependencies using APT repository.

- ✅ Multi-stage build pattern
- ✅ Many dependencies (build + runtime)
- ✅ Optimized production images
- ✅ Clear separation of concerns

**Best for:** Production applications, complex builds, optimized deployments

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

## Repository Types Comparison

| Type | Syntax | GPG Key | Use Case |
|------|--------|---------|----------|
| **PPA** | `ppa ppa:user/repo` | ✅ Automatic | Ubuntu community packages |
| **Custom Repo** | `deb "url"` + `key url` | ⚠️ Manual | Official vendor repos (Docker, etc.) |

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
