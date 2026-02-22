# Reproducible Builds with Aptfile.lock

This example demonstrates the lock file workflow for fully reproducible package installations.

## The Problem

Without a lock file, `apt curl` installs whatever version is current in the APT index.
Two builds a week apart may get different patch versions, making debugging harder and
builds less reproducible.

## The Solution: Aptfile.lock

`apt-bundle` provides a two-phase workflow:

### Phase 1: Generate the lock file (on your dev machine or in CI)

```bash
# Install packages as usual
sudo apt-bundle install

# Snapshot the exact installed versions into Aptfile.lock
apt-bundle lock
```

Commit both `Aptfile` and `Aptfile.lock` to version control.

### Phase 2: Reproducible install (in Docker or on another machine)

```bash
# Install the exact versions from Aptfile.lock
sudo apt-bundle install --locked
```

If a package in `Aptfile.lock` is not available at that exact version, the install fails
loudly — rather than silently installing a different version.

## What's in This Example

- **`Aptfile`** — Declares dependencies; one package is pinned directly (`jq=1.6-2.1ubuntu3`)
  and others are floating. Comments remind you to regenerate the lock file after changes.
- **`Aptfile.lock`** — Committed snapshot of exact versions (`name=version`, one per line,
  sorted alphabetically).
- **`Dockerfile`** — Copies both files and runs `apt-bundle install --locked` so the image
  always gets exactly the versions recorded in the lock file.

## Updating the Lock File

When you add, remove, or upgrade a package in `Aptfile`:

```bash
# Install to pick up the changes
sudo apt-bundle install

# Re-snapshot
apt-bundle lock

# Commit both files together
git add Aptfile Aptfile.lock
git commit -m "chore: update dependencies"
```

## Usage

```bash
# Build the image
make build

# Run interactively
make run

# Verify installed versions
make test
```

## When to Use This Pattern

- **CI/CD pipelines** — Guarantee identical packages on every run
- **Team environments** — Every developer gets the same package versions
- **Production Docker images** — Eliminate version drift between builds
- **Audit requirements** — Keep an explicit, reviewable record of installed versions
