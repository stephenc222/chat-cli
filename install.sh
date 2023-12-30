#!/usr/bin/env bash

# Define the installation directory
install_dir="/usr/local/bin"
repo="stephenc222/chat-cli"
# Base URL for binary downloads
base_url="https://github.com/$repo/releases/download"


# Check for Windows OS (not supported)
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    echo "Windows OS is not supported."
    exit 1
fi

# Determine OS and Architecture
os=""
arch=$(uname -m)

case "$(uname -s)" in
    Darwin)
        os="darwin"
        ;;
    Linux)
        os="linux"
        ;;
    *)
        echo "Unsupported operating system."
        exit 1
        ;;
esac

case $arch in
    x86_64)
        arch="amd64"
        ;;
    arm64)
        arch="arm64"
        ;;
    *)
        echo "Unsupported architecture."
        exit 1
        ;;
esac

echo "Detected OS: $os, Arch: $arch"

# API call to fetch the latest release tag from GitHub

latest_release_tag=$(curl -L -H "Accept: application/vnd.github+json" \
     -H "X-GitHub-Api-Version: 2022-11-28" \
     "https://api.github.com/repos/${repo}/releases" | jq '[.[] | {tag_name, created_at}] | sort_by(.created_at) | last(.[]).tag_name')


# Exit if the curl command or jq parsing fails
if [ -z "$latest_release_tag" ]; then
    echo "Failed to fetch the latest release tag. Exiting."
    exit 1
fi

# Remove quotes from the release tag
latest_release_tag=${latest_release_tag//\"/}
# Replace dots with underscores in the release tag
version=${latest_release_tag//./_}

# Construct the binary URL
binary_url="${base_url}/${latest_release_tag}/chat-${os}-${arch}-${version}"


# Use wget to download the binary
echo "Downloading chat-cli..."
curl -fsSL -o "${install_dir}/chat" "$binary_url"

# Exit if download fails
if [ $? -ne 0 ]; then
    echo "Failed to download chat-cli. Exiting."
    exit 1
fi

echo "Download complete."

# Change permissions to make the binary executable
if sudo chmod +x "${install_dir}/chat"; then
    echo "Installation complete. You can now run 'chat' from the terminal."
else
    echo "Failed to set execute permission on the binary. Installation incomplete."
    exit 1
fi
