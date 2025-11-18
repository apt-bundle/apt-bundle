# Installing apt-bundle via install.sh Script

A simple example demonstrating the easiest way to install `apt-bundle` using the installation script.

## Installation Method

This example uses the `install.sh` script, which is the simplest installation method:

```dockerfile
RUN curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | bash
```

## What's Included

- Build tools: `build-essential`
- Version control: `git`
- Text editors: `vim`
- Utilities: `curl`, `wget`, `jq`, `htop`
- Network tools: `net-tools`

## Usage

```bash
# Build the image
make build

# Run interactively
make run

# Or use Docker directly
docker build -t via-install-sh .
docker run -it via-install-sh
```

## When to Use This Method

- **Quick prototyping** - Fastest way to get started
- **Development environments** - Simple one-liner installation
- **Scripts and automation** - Easy to add to existing scripts
- **When you don't need version pinning** - Always gets the latest release

## Comparison

See `Dockerfile.traditional` for the equivalent Dockerfile without apt-bundle. The traditional approach requires:
- Long, hard-to-maintain RUN commands with all packages listed inline
- Difficult to add or remove packages (requires editing Dockerfile)
- No easy way to share dependency lists across projects
- Harder to version control dependency changes

Using apt-bundle provides a cleaner, more maintainable solution where dependencies are declared in `Aptfile` and can be easily shared and version controlled.
