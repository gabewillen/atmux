#!/bin/sh
set -e

# Detect OS/Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

VERSION=${AMUX_VERSION:-latest}
# URL would point to R2 or GitHub Releases
# Placeholder URL
URL="https://example.com/amux/releases/${VERSION}/amux-${OS}-${ARCH}"

echo "Installing amux for ${OS}/${ARCH}..."
# curl -L -o amux $URL
# chmod +x amux
# sudo mv amux /usr/local/bin/

echo "Done (placeholder)."
