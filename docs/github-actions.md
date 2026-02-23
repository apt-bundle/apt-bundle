---
layout: default
title: GitHub Actions
nav_order: 3.5
---

# GitHub Actions

The [`apt-bundle/apt-bundle-action`](https://github.com/apt-bundle/apt-bundle-action) GitHub Action provides a first-class integration for using apt-bundle in GitHub workflows. It handles downloading and installing apt-bundle automatically, and includes built-in package caching to speed up repeated runs.

## Quick Start

Add a step to your workflow:

```yaml
- uses: apt-bundle/apt-bundle-action@v1
```

This reads `Aptfile` from the repository root, installs packages, and caches the downloaded `.deb` files for future runs.

## Full Example

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install system dependencies
        uses: apt-bundle/apt-bundle-action@v1

      - name: Run tests
        run: make test
```

## Inputs

| Input | Default | Description |
|---|---|---|
| `file` | `Aptfile` | Path to the Aptfile |
| `mode` | `install` | `install`, `install-no-update`, or `check` |
| `version` | `latest` | apt-bundle version to install |
| `cache` | `true` | Cache downloaded `.deb` packages |
| `cache-key-prefix` | `apt-bundle` | Prefix for the cache key |

## Outputs

| Output | Description |
|---|---|
| `cache-hit` | `true` if packages were restored from cache |

## Caching

Caching is enabled by default. The cache key is derived from a SHA256 hash of `Aptfile.lock` (if present) or `Aptfile`. The cache is automatically invalidated whenever dependencies change.

To disable caching:

```yaml
- uses: apt-bundle/apt-bundle-action@v1
  with:
    cache: false
```

## Reproducible Builds with Aptfile.lock

If your repository includes an `Aptfile.lock`, the action uses it as the cache key and installs the exact package versions recorded in the lockfile, ensuring reproducible builds across runs and machines.

```yaml
- uses: apt-bundle/apt-bundle-action@v1
# Automatically uses Aptfile.lock if present
```

See the [Aptfile Format](aptfile-format.html) page for how to generate a lockfile.

## Validating Dependencies in Pull Requests

Use `mode: check` to verify that the Aptfile is consistent with what the build environment provides, without installing anything:

```yaml
- name: Validate dependencies
  uses: apt-bundle/apt-bundle-action@v1
  with:
    mode: check
```

This is useful as a lightweight lint step on pull requests.

## Pinning a Version

To use a specific apt-bundle release rather than `latest`:

```yaml
- uses: apt-bundle/apt-bundle-action@v1
  with:
    version: 1.2.3
```

## Custom Aptfile Location

```yaml
- uses: apt-bundle/apt-bundle-action@v1
  with:
    file: .github/Aptfile
```

## Requirements

- Ubuntu runners only (`ubuntu-*`)
- Standard GitHub-hosted runners have the necessary `sudo` access

---

See the [action repository](https://github.com/apt-bundle/apt-bundle-action) for the full source and changelog.
