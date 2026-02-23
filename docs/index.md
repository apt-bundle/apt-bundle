---
layout: default
title: Home
nav_order: 1
---

# apt-bundle

A declarative package manager for apt, inspired by Homebrew's `brew bundle`.

## Overview

`apt-bundle` provides a simple, declarative, and shareable way to manage apt packages and repositories on Debian-based systems. Define your system dependencies in an `Aptfile` and install them with a single command.

## Features

- **Declarative Package Management**: Define packages in a simple text file
- **Idempotent Operations**: Safe to run multiple times
- **Repository & Key Management**: Add PPAs, custom repositories, and GPG keys
- **Version Pinning**: Install specific package versions
- **Simple CLI**: Easy-to-use command-line interface
- **GitHub Actions Integration**: Native action with built-in caching for CI/CD

## Quick Start

```bash
# Clone the repository
git clone https://github.com/apt-bundle/apt-bundle.git
cd apt-bundle

# Build and install
make build
sudo make install

# Create an Aptfile
cat > Aptfile <<EOF
apt vim
apt curl
apt git
EOF

# Install packages
sudo apt-bundle
```

## Use Cases

### Developer Onboarding
A new developer joins a project, clones the repo, and runs `sudo apt-bundle` to get all required system dependencies.

### Dockerfile Build
Replace long, unmaintainable `RUN apt-get install -y ...` lines with a simple `Aptfile` and `apt-bundle` command.

### System Sync
Use `apt-bundle dump > Aptfile` on your primary workstation and then `sudo apt-bundle` on a new laptop to sync your tools.

### CI/CD
Use the [GitHub Action](github-actions.html) for seamless integration with GitHub workflows, including built-in package caching and reproducible builds via lockfiles.

## Documentation

- [Installation](installation.html) - How to install apt-bundle
- [Usage](usage.html) - Command reference and examples
- [GitHub Actions](github-actions.html) - Using apt-bundle in GitHub workflows
- [Aptfile Format](aptfile-format.html) - Complete syntax reference
- [Contributing](contributing.html) - Contributing and development setup
- [API Reference](api-reference.html) - Go package documentation

