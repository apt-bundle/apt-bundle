#!/bin/bash
# Script to update the APT repository with new releases
# Usage: ./update-repo.sh [VERSION]
# Example: ./update-repo.sh 0.1.8
#
# This script works on both Linux (with apt-utils installed) and macOS (using Docker)

set -euo pipefail

VERSION="${1:-}"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POOL_DIR="${REPO_DIR}/pool/main/a/apt-bundle"
DISTS_DIR="${REPO_DIR}/dists/stable"

# Detect if we need to use Docker for apt tools
USE_DOCKER=false
if ! command -v dpkg-scanpackages &> /dev/null || ! command -v apt-ftparchive &> /dev/null; then
    echo "Note: apt-utils not found locally, will use Docker"
    USE_DOCKER=true
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is required on macOS/non-Debian systems"
        exit 1
    fi
fi

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: $0 <VERSION>"
    echo "Example: $0 0.1.8"
    exit 1
fi

echo "Updating APT repository for version ${VERSION}..."

# Download packages for all architectures
for arch in amd64 arm64 armhf i386; do
    PACKAGE_NAME="apt-bundle_${VERSION}_linux_${arch}.deb"
    PACKAGE_URL="https://github.com/apt-bundle/apt-bundle/releases/download/v${VERSION}/${PACKAGE_NAME}"
    
    echo "Downloading ${PACKAGE_NAME}..."
    curl -L -o "${POOL_DIR}/${PACKAGE_NAME}" "${PACKAGE_URL}"
done

# Regenerate Packages files for each architecture
echo "Generating Packages files..."
if [ "$USE_DOCKER" = true ]; then
    for arch in amd64 arm64 armhf i386; do
        echo "  - ${arch}"
        docker run --rm -v "${REPO_DIR}:/repo" -w /repo ubuntu:22.04 bash -c \
            "apt-get update -qq > /dev/null 2>&1 && apt-get install -y -qq dpkg-dev > /dev/null 2>&1 && \
             dpkg-scanpackages --arch ${arch} pool/main 2>/dev/null" | \
            grep -v "^dpkg-scanpackages:" | \
            sed 's|^Filename: .*/pool/|Filename: pool/|' > "${DISTS_DIR}/main/binary-${arch}/Packages"
        gzip -k -f "${DISTS_DIR}/main/binary-${arch}/Packages"
    done
else
    for arch in amd64 arm64 armhf i386; do
        echo "  - ${arch}"
        (cd "${REPO_DIR}" && dpkg-scanpackages --arch "${arch}" pool/main 2>/dev/null) | \
            grep -v "^dpkg-scanpackages:" | \
            sed 's|^Filename: .*/pool/|Filename: pool/|' > "${DISTS_DIR}/main/binary-${arch}/Packages"
        gzip -k -f "${DISTS_DIR}/main/binary-${arch}/Packages"
    done
fi

# Regenerate Release file
echo "Generating Release file..."
if [ "$USE_DOCKER" = true ]; then
    docker run --rm -v "${REPO_DIR}:/repo" -w /repo ubuntu:22.04 bash -c \
        "apt-get update -qq > /dev/null 2>&1 && apt-get install -y -qq apt-utils > /dev/null 2>&1 && \
         apt-ftparchive -c Release.conf release dists/stable" > "${DISTS_DIR}/Release"
else
    apt-ftparchive -c "${REPO_DIR}/Release.conf" release "${DISTS_DIR}" > "${DISTS_DIR}/Release"
fi

echo "✅ Repository updated successfully!"
echo ""
echo "Repository structure:"
tree -L 3 "${REPO_DIR}" || find "${REPO_DIR}" -type f | head -20

