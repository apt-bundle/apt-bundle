# CI/CD Build Image

Docker image for CI/CD pipelines with build tools, testing frameworks, Docker CLI, and cloud provider tools.

This example demonstrates installing `apt-bundle` from the official APT repository at `https://apt-bundle.org/repo/`.

## What's Included

- Build tools: `build-essential`, `cmake`, `pkg-config`
- Version control: `git`, `git-lfs`
- Testing: `shellcheck`, `jq`, `curl`
- Docker: `docker-ce-cli`, `docker-buildx-plugin` (for Docker-in-Docker)
- Cloud CLIs: `awscli`, `azure-cli`

## Usage

```bash
# Build the image
make build

# Run with Docker socket (for Docker-in-Docker)
make run

# Run in CI mode
make ci
```

## Docker-in-Docker

To use Docker commands inside the container, mount the Docker socket:
```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock ci-builder
```

## Comparison

See `Dockerfile.traditional` for the equivalent Dockerfile without apt-bundle. The traditional approach requires:
- Long RUN commands with all packages listed inline
- Manual GPG key management and repository setup for Docker
- Complex multi-step process to add custom repositories
- Difficult to maintain and update dependencies
- Hard to share CI/CD dependency lists across projects

Using apt-bundle simplifies this significantly by handling repository setup, GPG keys, and package installation declaratively through the `Aptfile`.

## Installing apt-bundle

This example installs `apt-bundle` from the official APT repository:

```dockerfile
RUN echo "deb [arch=amd64,arm64,armhf,i386] [trusted=yes] https://apt-bundle.org/repo/ stable main" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y apt-bundle
```

Note: The `trusted=yes` flag is used because the repository is unsigned. For production use, consider setting up GPG signing for the repository.

Alternatively, you can use the installation script:
```dockerfile
RUN curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | bash
```

