#!/bin/bash

# Define platforms and architectures
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64" "windows/arm64")

# Create build directory
mkdir -p build

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name="build/partition-cleaner-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi

    echo "Building for $GOOS/$GOARCH..."
    
    # Get version from git tag or commit
    VERSION=$(git describe --tags --always --dirty)
    
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.version=$VERSION" -o $output_name
    if [ $? -ne 0 ]; then
        echo "An error has occurred! Aborting the script execution..."
        exit 1
    fi
done

echo "Build complete! Binaries are in the build/ directory."
