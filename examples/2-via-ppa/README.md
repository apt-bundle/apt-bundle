# Using PPAs (Personal Package Archives)

A practical example demonstrating how to use `apt-bundle` with PPAs (Personal Package Archives) to install packages from community-maintained repositories.

## What Makes PPAs Special

PPAs are **simpler than custom repositories** because:
- ✅ **No GPG key management** - PPAs automatically handle GPG keys
- ✅ **Simple syntax** - Just `ppa ppa:user/repo`
- ✅ **Ubuntu integration** - Built-in support via `add-apt-repository`
- ✅ **Common use case** - Many popular packages are available via PPAs

## Installation Method

This example uses the `install.sh` script (PPAs work with either installation method):

```dockerfile
RUN curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | bash
```

## What's Included

This example demonstrates installing packages from multiple PPAs:

- **PHP 8.2** from `ppa:ondrej/php` (popular PHP PPA)
- **Git** from `ppa:git-core/ppa` (newer Git versions)
- **Node.js** from `ppa:chris-lea/node.js` (Node.js packages)
- **Standard packages** from Ubuntu repos (build-essential, curl, vim)

## Usage

```bash
# Build the image
make build

# Run interactively
make run

# Or use Docker directly
docker build -t via-ppa .
docker run -it via-ppa
```

## PPA vs Custom Repository Comparison

### PPAs (This Example)
```aptfile
# Simple - no GPG key needed!
ppa ppa:ondrej/php
apt php8.2
```

### Custom Repositories (Example 3)
```aptfile
# More complex - requires GPG key
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu jammy stable"
apt docker-ce
```

## When to Use PPAs

- ✅ **Ubuntu/Debian packages** - PPAs are Ubuntu-specific
- ✅ **Community packages** - Popular packages maintained by community
- ✅ **Newer versions** - Get latest versions not in official repos
- ✅ **Simple setup** - No GPG key management needed

## Common PPAs

Some popular PPAs you might use:

- `ppa:ondrej/php` - PHP versions (5.6 through 8.3)
- `ppa:git-core/ppa` - Latest Git versions
- `ppa:chris-lea/node.js` - Node.js packages
- `ppa:deadsnakes/ppa` - Python versions for Ubuntu
- `ppa:ubuntu-toolchain-r/test` - GCC toolchain versions

## Comparison

See `Dockerfile.traditional` for the equivalent Dockerfile without apt-bundle. The traditional approach requires:
- Manual `add-apt-repository` commands for each PPA
- Remembering to run `apt-get update` after adding PPAs
- Hard to track which packages come from which PPA
- Difficult to share PPA setups across projects

Using apt-bundle with PPAs provides a clean, declarative way to manage PPA-based packages that can be easily shared and version controlled.

