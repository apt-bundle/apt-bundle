# Example 6: GitHub CLI

This example demonstrates installing **GitHub CLI** (`gh`) from its official APT repository
using apt-bundle's `key` + `deb` pattern — the most common real-world use case for custom
third-party repositories with GPG key pinning.

## The Aptfile

```aptfile
# GitHub CLI - https://cli.github.com
# 'key' downloads and saves the GPG key; the following 'deb' is automatically
# signed by it (Signed-By is set in the generated DEB822 sources file).
key https://cli.github.com/packages/githubcli-archive-keyring.gpg
deb "[arch=amd64] https://cli.github.com/packages stable main"
apt gh
```

apt-bundle automatically links the downloaded GPG key to the repository: no need to
spell out `signed-by=` in the `deb` line.

## Explicit `signed-by=` (copy-paste from official docs)

Official GitHub CLI installation instructions include `signed-by=` in the deb options.
apt-bundle now accepts that form too, so you can copy-paste it directly:

```aptfile
key https://cli.github.com/packages/githubcli-archive-keyring.gpg
deb "[arch=amd64 signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main"
apt gh
```

When `signed-by=` is present in the `deb` options it takes precedence over the
implicit key path — the path must already exist on the system (written by the
preceding `key` directive).

## Usage

```bash
# Build with apt-bundle (Dockerfile)
make build

# Or build the traditional equivalent (no apt-bundle)
make build DOCKERFILE=Dockerfile.traditional

# Test: runs gh --version inside the container
make test

# Interactive shell
make shell
```

## What `Dockerfile.traditional` shows

`Dockerfile.traditional` is the equivalent ~15-line shell block that the three
apt-bundle lines replace: download key with curl, dearmor, write sources file,
`apt-get update`, `apt-get install gh`. It is provided for comparison only.
