---
layout: default
title: Aptfile Format
nav_order: 4
---

# Aptfile Format

The `Aptfile` is a simple, line-oriented, UTF-8 encoded text file. The default filename is `Aptfile`, but you can specify a different path using the `--file` flag.

## Basic Rules

- **Comments**: Lines beginning with `#` are ignored
- **Empty Lines**: Empty lines are ignored
- **Directives**: Each line consists of a directive (keyword) followed by an argument
- **Quotes**: Quotes (`"` or `'`) are required if the argument contains spaces or special characters

## Directives

### `apt` - Install Packages

Specifies a standard apt package to be installed.

**Syntax:** `apt <package-name>[=<version-string>]`

**Logic:** Corresponds to `apt-get install -y <package-name>[=<version-string>]`

**Examples:**

```aptfile
# Install latest version
apt vim
apt curl
apt build-essential

# Install specific version
apt "nano=2.9.3-2"
apt "docker-ce=5:19.03.13~3-0~ubuntu-focal"
```

**Notes:**
- Package names are case-sensitive
- Version strings must be quoted if they contain special characters
- If a version is specified, that exact version will be installed
- If no version is specified, the latest available version will be installed

### `ppa` - Add Personal Package Archive

Specifies a Personal Package Archive (PPA) to be added.

**Syntax:** `ppa <ppa:user/repo>`

**Logic:** Corresponds to `add-apt-repository -y <ppa:user/repo>`

**Examples:**

```aptfile
ppa ppa:ondrej/php
ppa ppa:git-core/ppa
```

**Notes:**
- This directive implicitly handles adding the PPA's GPG key
- PPAs must be in the format `ppa:user/repo`
- PPAs are typically Ubuntu-specific

### `deb` - Add Custom Repository

Specifies a full deb repository line to be added to `/etc/apt/sources.list.d/`.

**Syntax:** `deb "<full-repository-line>"`

**Logic:** Creates a `.list` file in `/etc/apt/sources.list.d/` containing the repository line.

**Examples:**

```aptfile
# Google Chrome
deb "[arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main"

# Docker
deb "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
```

**Notes:**
- The repository line must be quoted (usually with double quotes)
- This directive only adds the repository line; it does not add the GPG key
- You must use the `key` directive separately to add the required GPG key
- Repository lines can include options like `[arch=amd64]` or `[trusted=yes]`
- `signed-by=<path>` is supported in the options block and takes precedence over the implicit key linking from a preceding `key` directive; this is useful for copy-paste compatibility with official installation docs

**Recommended pattern** — use `key` + `deb` without `signed-by=`; apt-bundle automatically links them:

```aptfile
key https://cli.github.com/packages/githubcli-archive-keyring.gpg
deb "[arch=amd64] https://cli.github.com/packages stable main"
apt gh
```

**Advanced / copy-paste from official docs** — explicit `signed-by=` is also accepted:

```aptfile
key https://cli.github.com/packages/githubcli-archive-keyring.gpg
deb "[arch=amd64 signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main"
apt gh
```

When using the explicit form, the path in `signed-by=` must already exist on the system (e.g. written by the `key` directive to `/etc/apt/keyrings/`).

### `key` - Add GPG Key

Specifies a GPG key URL to be downloaded and added to `/etc/apt/trusted.gpg.d/`.

**Syntax:** `key <url-to-key>`

**Logic:** Downloads the key, dearmors it, and saves it to `/etc/apt/trusted.gpg.d/<filename>.gpg`.

**Examples:**

```aptfile
# Google Chrome Key
key https://dl.google.com/linux/linux_signing_key.pub

# Docker Key
key https://download.docker.com/linux/ubuntu/gpg
```

**Notes:**
- Keys are typically required for custom repositories (`deb` directives)
- Keys are not required for PPAs (handled automatically)
- The URL can point to either an armored (`.pub`, `.asc`) or binary key file
- Keys must be added before the repository that uses them

## Complete Example

Here's a complete example `Aptfile` demonstrating all directives:

```aptfile
# This is a sample Aptfile

# Core development tools
apt build-essential
apt curl
apt git
apt vim
apt htop

# Specific version of nano
apt "nano=2.9.3-2"

# PHP from a PPA
ppa ppa:ondrej/php
apt php8.1
apt php8.1-cli
apt php8.1-fpm

# Install Docker
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
apt docker-ce
apt docker-ce-cli
apt containerd.io
```

## Order of Operations

When `apt-bundle install` runs, it processes directives in the following order:

1. All `key` directives (add GPG keys)
2. All `ppa` directives (add PPAs)
3. All `deb` directives (add repositories)
4. Run `apt-get update` (by default, can be skipped with `--no-update` flag)
5. All `apt` directives (install packages)

This ensures that repositories and keys are set up before attempting to install packages. The package list update runs by default to ensure fresh package information, but can be skipped using the `--no-update` flag if your package lists are already up-to-date (e.g., in CI/CD environments).

## Best Practices

1. **Group related items**: Keep related packages, repositories, and keys together
2. **Use comments**: Add comments to explain why certain packages or repositories are needed
3. **Pin versions carefully**: Only pin versions when necessary for reproducibility
4. **Order matters**: Place `key` directives before `deb` directives that use them
5. **Test your Aptfile**: Use `apt-bundle check` before running `apt-bundle install`

## Common Patterns

### Development Environment

```aptfile
# Build tools
apt build-essential
apt cmake
apt pkg-config

# Version control
apt git
apt git-lfs

# Editors
apt vim
apt nano
```

### Docker Environment

```aptfile
# Docker repository
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Docker packages
apt docker-ce
apt docker-ce-cli
apt containerd.io
apt docker-compose-plugin
```

### GitHub CLI

```aptfile
key https://cli.github.com/packages/githubcli-archive-keyring.gpg
deb "[arch=amd64] https://cli.github.com/packages stable main"
apt gh
```

The `signed-by=` form from the official GitHub CLI docs is also accepted if you prefer
to copy-paste it directly.

### Language-Specific Setup

```aptfile
# PHP from PPA
ppa ppa:ondrej/php
apt php8.1
apt php8.1-cli
apt php8.1-fpm
apt php8.1-mysql

# Node.js repository
key https://deb.nodesource.com/gpgkey/nodesource.gpg.key
deb "https://deb.nodesource.com/node_18.x $(lsb_release -cs) main"
apt nodejs
```

## Related Documentation

- [Usage Guide](usage.html) - How to use apt-bundle commands
- [Technical Specification](https://github.com/apt-bundle/apt-bundle/blob/main/specs/tech-specs.md) - Internal specification
- [Requirements](https://github.com/apt-bundle/apt-bundle/blob/main/specs/requirements.md) - Internal specification

