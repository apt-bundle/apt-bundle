# Complex Multi-Stage Build with APT Repository

A comprehensive example demonstrating advanced usage of `apt-bundle` with multi-stage builds, many dependencies, and the APT repository installation method.

## Installation Method

This example uses the official APT repository in both build and runtime stages:

```dockerfile
RUN echo "deb [arch=amd64,arm64,armhf,i386] [trusted=yes] https://apt-bundle.org/repo/ stable main" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y apt-bundle
```

## What Makes This Complex

### Multi-Stage Build
- **Build stage**: Contains all development tools and libraries needed to compile the application
- **Runtime stage**: Contains only runtime libraries needed to run the application
- **Result**: Smaller, more secure production images

### Many Dependencies

**Build Stage (`Aptfile.build`):**
- Compiler toolchain: `build-essential`, `cmake`, `autoconf`, `automake`
- Development libraries: `libssl-dev`, `libffi-dev`, `libpq-dev`, `libmysqlclient-dev`
- Image processing dev libs: `libjpeg-dev`, `libpng-dev`, `libtiff-dev`
- Source tools: `git`, `git-lfs`, `curl`, `wget`

**Runtime Stage (`Aptfile.runtime`):**
- Runtime libraries: `libssl3`, `libffi8`, `libpq5`, `libmysqlclient21`
- Image processing runtime: `libjpeg8`, `libpng16-16`, `libtiff5`
- System libraries: `ca-certificates`, `zlib1g`, `liblzma5`

## Usage

```bash
# Build the image
make build

# Run the application
make run

# Inspect image size (compare build vs runtime)
make size
```

## Benefits

- ✅ **Smaller final image** - Build tools excluded from production
- ✅ **Better security** - Fewer attack surfaces in runtime image
- ✅ **Faster deployments** - Smaller images deploy faster
- ✅ **Clear separation** - Build vs runtime dependencies are explicit
- ✅ **Version control** - Both Aptfiles can be version controlled separately

## When to Use This Pattern

- **Production applications** - Need optimized, secure images
- **Complex builds** - Applications requiring many build dependencies
- **Multi-architecture** - Building for different platforms
- **CI/CD pipelines** - Separate build and deploy stages

## Comparison

See `Dockerfile.traditional` for the equivalent Dockerfile without apt-bundle. The traditional approach requires:
- Duplicating package lists in both build and runtime stages
- Manual management of which packages belong in which stage
- Harder to maintain consistency between stages
- Difficult to share dependency lists
- Easy to accidentally include build tools in runtime stage

Using apt-bundle with separate `Aptfile.build` and `Aptfile.runtime` files makes it clear which dependencies belong to which stage and allows easy sharing and version control.
