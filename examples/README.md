# Dockerfile Examples with Aptfile

This directory contains practical examples demonstrating different ways to use `apt-bundle` with Dockerfiles.

## Examples Overview

The examples are organized by **installation method** and **complexity**, not by package types:

### 1. Via install.sh Script (`1-via-install-sh/`)
**Simplest approach** - Quick installation using the install script.

- ✅ Fastest to get started
- ✅ One-liner installation
- ✅ Always gets latest version
- ⚠️ No version pinning

**Best for:** Quick prototyping, development environments, scripts

### 2. Via APT Repository (`2-via-apt-get/`)
**Production-ready approach** - Install from official APT repository.

- ✅ Version control
- ✅ Better caching
- ✅ Multi-architecture support
- ✅ Production-ready

**Best for:** Production Docker images, CI/CD pipelines, when you need specific versions

### 3. Complex via APT Repository (`3-complex-via-apt-get/`)
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

### Example 2: APT repository approach
```bash
cd 2-via-apt-get
make build
make run
```

### Example 3: Complex multi-stage
```bash
cd 3-complex-via-apt-get
make build
make run
```

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
- Each example can be customized to fit your specific needs
