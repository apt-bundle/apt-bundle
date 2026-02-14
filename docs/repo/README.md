---
title: APT Repository
---

# apt-bundle APT Repository

This directory contains the APT repository for the `apt-bundle` package.

## Repository URL

```
https://apt-bundle.org/repo/
```

## Adding the Repository

To add this repository to your system:

```bash
echo "deb [arch=amd64,arm64,armhf,i386] https://apt-bundle.org/repo/ stable main" | sudo tee /etc/apt/sources.list.d/apt-bundle.list
sudo apt update
sudo apt install apt-bundle
```

Or for a specific architecture:

```bash
# For amd64
echo "deb [arch=amd64] https://apt-bundle.org/repo/ stable main" | sudo tee /etc/apt/sources.list.d/apt-bundle.list

# For arm64
echo "deb [arch=arm64] https://apt-bundle.org/repo/ stable main" | sudo tee /etc/apt/sources.list.d/apt-bundle.list

# For armhf
echo "deb [arch=armhf] https://apt-bundle.org/repo/ stable main" | sudo tee /etc/apt/sources.list.d/apt-bundle.list

# For i386
echo "deb [arch=i386] https://apt-bundle.org/repo/ stable main" | sudo tee /etc/apt/sources.list.d/apt-bundle.list
```

## Updating the Repository

To update the repository with a new release version:

```bash
cd docs/repo
./update-repo.sh <VERSION>
```

Example:
```bash
./update-repo.sh 0.1.9
```

## Repository Structure

```
repo/
├── dists/
│   └── stable/
│       ├── Release
│       └── main/
│           └── binary-{arch}/
│               ├── Packages
│               └── Packages.gz
├── pool/
│   └── main/
│       └── a/
│           └── apt-bundle/
│               └── apt-bundle_{VERSION}_linux_{arch}.deb
├── Release.conf
├── update-repo.sh
└── README.md
```

## Supported Architectures

- `amd64` (x86_64)
- `arm64` (aarch64)
- `armhf` (ARMv7 hard-float)
- `i386` (x86)

## Current Version

The repository currently contains version **{{ site.data.version.apt_repo_version }}**.

## Notes

- This repository is served via GitHub Pages
- The repository is not GPG signed (unsigned repository)
- To use an unsigned repository, you may need to add `[trusted=yes]` to the repository line or configure APT to allow unsigned repositories

