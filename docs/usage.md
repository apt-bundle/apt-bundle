---
layout: default
title: Usage
nav_order: 3
---

# Usage

## Basic Commands

### Install (Default Command)

Install packages from an `Aptfile`:

```bash
# Use default Aptfile (./Aptfile)
sudo apt-bundle

# Or explicitly
sudo apt-bundle install

# Use a different Aptfile
sudo apt-bundle --file /path/to/Aptfile
sudo apt-bundle install --file /path/to/Aptfile
```

The `install` command:
1. Adds all specified repositories and GPG keys
2. Runs `apt-get update` (by default, can be skipped with `--no-update`)
3. Installs all specified packages

**Note:** The `install` command requires root privileges (use `sudo`).

### Check

Check if packages and repositories from the Aptfile are installed (no root required):

```bash
# Check using default Aptfile
apt-bundle check

# Check using a different file
apt-bundle check --file /path/to/Aptfile
```

This command reads the Aptfile and reports which packages/repositories are missing or not installed, without actually installing them.

### Dump

Generate an Aptfile from the system's current state:

```bash
# Output to stdout
apt-bundle dump

# Save to file
apt-bundle dump > Aptfile
```

This command outputs a list of manually installed packages. Future versions may also include custom PPAs and deb repositories.

## Command-Line Options

### Global Flags

- `--file, -f`: Specify the path to the Aptfile (default: `./Aptfile`)
- `--help, -h`: Show help information
- `--version`: Show version information

### Install Command Flags

- `--no-update`: Skip updating package lists before installing packages. By default, `apt-bundle` runs `apt-get update` to ensure fresh package lists.

## Examples

### Basic Package Installation

Create an `Aptfile`:

```aptfile
apt vim
apt curl
apt git
apt build-essential
```

Install:

```bash
sudo apt-bundle
```

### Using a Custom Aptfile Location

```bash
sudo apt-bundle --file /etc/myproject/Aptfile
```

### Skipping Package List Update

If your package lists are already up-to-date (e.g., in CI/CD where you've already run `apt-get update`):

```bash
sudo apt-bundle --no-update
```

### Checking Before Installing

```bash
# Check what's missing
apt-bundle check

# Review the output, then install
sudo apt-bundle
```

### Generating an Aptfile

```bash
# On your primary workstation
apt-bundle dump > Aptfile

# Commit to version control
git add Aptfile
git commit -m "Add system dependencies"

# On a new machine
git clone <repo>
cd <repo>
sudo apt-bundle
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

Use `apt-bundle` in your Dockerfiles to manage system dependencies declaratively. See the [examples directory](../../examples/) for complete working examples including multi-stage builds, Python runtimes, CI/CD setups, and more.

### CI/CD with GitHub Actions

The [`apt-bundle/apt-bundle-action`](https://github.com/apt-bundle/apt-bundle-action) GitHub Action provides a streamlined integration with built-in package caching:

```yaml
# .github/workflows/test.yml
- name: Install system dependencies
  uses: apt-bundle/apt-bundle-action@v1
```

To validate dependencies without installing (e.g. in a lint step):

```yaml
- name: Check dependencies
  uses: apt-bundle/apt-bundle-action@v1
  with:
    mode: check
```

See the [GitHub Actions](github-actions.html) page for the full reference.

### System Sync

```bash
# On primary workstation
apt-bundle dump > ~/dotfiles/Aptfile

# On new laptop
cp ~/dotfiles/Aptfile .
sudo apt-bundle
```

## Idempotency

All operations are idempotent, meaning you can safely run `apt-bundle install` multiple times:

- Packages already installed will be skipped
- Repositories already added will not be duplicated
- GPG keys already present will not be re-added

This makes `apt-bundle` safe to run repeatedly, which is especially useful in CI/CD pipelines and Dockerfiles.

## Error Handling

If `apt-bundle` encounters an error:

- It will exit with a non-zero status code
- Error messages will be displayed clearly
- Partial installations will not leave the system in an inconsistent state

Common errors:
- Package not found: Check package name spelling
- Invalid PPA: Verify PPA format (`ppa:user/repo`)
- Permission denied: Ensure you're using `sudo` for install command
- Network errors: Check internet connectivity

## Next Steps

- Learn about the [Aptfile Format](aptfile-format.html) for detailed syntax
- See [Installation](installation.html) if you haven't installed yet
- Check out the [Contributing](contributing.html) guide to contribute

