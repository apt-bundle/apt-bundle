#!/bin/bash
# Script to update the APT repository with new releases
# Usage: ./update-repo.sh [VERSION]
# Example: ./update-repo.sh 0.1.8

set -euo pipefail

VERSION="${1:-}"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POOL_DIR="${REPO_DIR}/pool/main/a/apt-bundle"
DISTS_DIR="${REPO_DIR}/dists/stable"

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
for arch in amd64 arm64 armhf i386; do
    echo "  - ${arch}"
    dpkg-scanpackages --arch "${arch}" "${REPO_DIR}/pool/main" > "${DISTS_DIR}/main/binary-${arch}/Packages" 2>/dev/null
    gzip -k -f "${DISTS_DIR}/main/binary-${arch}/Packages"
done

# Regenerate Release file
echo "Generating Release file..."
apt-ftparchive -c "${REPO_DIR}/Release.conf" release "${DISTS_DIR}" > "${DISTS_DIR}/Release"

echo "✅ Repository updated successfully!"
echo ""
echo "Repository structure:"
tree -L 3 "${REPO_DIR}" || find "${REPO_DIR}" -type f | head -20

