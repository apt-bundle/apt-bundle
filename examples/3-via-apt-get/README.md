# Installing apt-bundle via APT Repository

A production-ready example demonstrating how to install `apt-bundle` from the official APT
repository, including adding a custom third-party repository (Docker) via `key` and `deb`
directives.

## Installation Method

This example uses the official APT repository, which provides better control and version management:

```dockerfile
RUN echo "deb [arch=amd64,arm64,armhf,i386 trusted=yes] https://apt-bundle.org/repo/ stable main" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y apt-bundle
```

## What's Included

- Standard utilities: `curl`, `git`, `jq`
- Docker CLI — installed from Docker's official repository via `key` + `deb` directives

## The key + deb Pattern

The `key` directive downloads a GPG key over HTTPS and the immediately following `deb`
directive adds the repository with that key as its `Signed-By` field:

```aptfile
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu jammy stable"
apt docker-ce-cli
```

This replaces around 10 lines of manual shell commands (curl, gpg --dearmor, chmod, echo,
tee, etc.) that you would otherwise need in a `Dockerfile`. See `Dockerfile.traditional`
for the equivalent without apt-bundle.

## Usage

```bash
# Build the image
make build

# Run interactively
make run

# Test the installed packages
make test
```

## When to Use This Method

- **Production environments** — More reliable and version-controlled
- **Docker images** — Better layer caching and reproducibility
- **CI/CD pipelines** — Consistent versions across builds
- **Custom third-party repositories** — `key` + `deb` handles GPG key download and
  DEB822 repository setup automatically
- **Multi-architecture support** — Supports amd64, arm64, armhf, i386

## Benefits Over install.sh

- ✅ Version control — Can pin to specific versions
- ✅ Better caching — APT caches packages efficiently
- ✅ Multi-architecture — Supports all architectures
- ✅ Production-ready — Standard APT package management
- ✅ Reproducible — Same version every time
- ✅ Custom repos — `key` + `deb` directives manage GPG keys and sources automatically

## Note

The `trusted=yes` flag is used in the apt-bundle repository entry because that repository
is currently unsigned. The Docker repository itself uses proper GPG signing via the `key`
directive in the Aptfile.
