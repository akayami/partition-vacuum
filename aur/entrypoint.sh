#!/bin/bash
set -e

# If a version argument is passed, use it. Otherwise try to detect from git.
VERSION="$1"

if [ -z "$VERSION" ]; then
    # Fallback if no version provided, though in CI we should provide it.
    # Inside the container, .git might be present if copied.
    if [ -d ".git" ]; then
        VERSION=$(git describe --tags --always | sed 's/^v//;s/-/+/g')
    else
        echo "Error: No version provided and not a git repository."
        exit 1
    fi
fi

echo "Building version: $VERSION"

# Update PKGBUILD version
sed -i "s/^pkgver=.*/pkgver=${VERSION}/" aur/PKGBUILD

# Move PKGBUILD to root of workdir for makepkg
cp aur/PKGBUILD .

# Update checksums (though we use SKIP for local source, good practice if we change source later)
updpkgsums

# Build the package
makepkg --printsrcinfo > .SRCINFO
makepkg -s --noconfirm

# Output the filename for easy retrieval
ls -1 *.pkg.tar.zst

# If /output directory exists, copy artifacts there
if [ -d "/output" ]; then
    echo "Copying artifacts to /output..."
    sudo cp *.pkg.tar.zst /output/
    # Also copy PKGBUILD and .SRCINFO for reference
    sudo cp PKGBUILD .SRCINFO /output/
    sudo chown -R $(stat -c "%u:%g" /output) /output
fi
