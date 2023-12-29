#!/usr/bin/env bash

# Define target platforms and architectures
PLATFORMS=("darwin" "linux") # darwin is macOS
ARCHS=("amd64" "arm64")      # amd64 is x86, arm64 is ARM

LATEST_TAG=$(git describe --abbrev=0 --tags)
LATEST_TAG=${LATEST_TAG//./_}

if [ -z "$LATEST_TAG" ]; then
  echo "No Git tags found. Cannot start build, exiting"
  exit 1
fi

# Check the Go version
required_go_version="1.20.0"  # Minimum required version

installed_go_version=$(go version | awk '{print $3}' | sed 's/^go//')

echo $installed_go_version

if ! [[ $installed_go_version =~ ^[0-9]+\.[0-9]+\.[0-9] ]]; then
    echo "Unable to determine Go version"
    exit 1
fi

if [[ $installed_go_version < $required_go_version ]]; then
    echo "Go version $required_go_version or higher is required"
    exit 1
fi

echo "Go version $required_go_version or higher is installed, starting build..."


# Directory to store binaries
BIN_DIR="./bin"
mkdir -p $BIN_DIR

go mod tidy

# Function to build for each platform and architecture
build() {
    local os=$1
    local arch=$2
    local output_name="chat-${os}-${arch}-${LATEST_TAG}"

    echo "Building for OS: $os, Arch: $arch, version: $LATEST_TAG"
    GOOS=$os GOARCH=$arch go build -o "${BIN_DIR}/${output_name}" .
}

# Build for each combination of platform and architecture
for platform in "${PLATFORMS[@]}"; do
    for arch in "${ARCHS[@]}"; do
        build $platform $arch
    done
done

echo "Build complete."