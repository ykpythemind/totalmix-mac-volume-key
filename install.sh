#!/bin/bash
set -euo pipefail

REPO="ykpythemind/totalmix-mac-volume-key"
INSTALL_DIR="/usr/local/bin"
BIN_NAME="totalmix-mac-volume-key"

ARCH=$(uname -m)
case "$ARCH" in
  arm64) ASSET="${BIN_NAME}-darwin-arm64" ;;
  x86_64) ASSET="${BIN_NAME}-darwin-amd64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Fetching latest release..."
DOWNLOAD_URL=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep "browser_download_url.*${ASSET}" \
  | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Error: Could not find download URL for ${ASSET}"
  exit 1
fi

TMPFILE=$(mktemp)
echo "Downloading ${ASSET}..."
curl -sL -o "$TMPFILE" "$DOWNLOAD_URL"
chmod +x "$TMPFILE"

echo "Installing to ${INSTALL_DIR}/${BIN_NAME} (requires sudo)..."
sudo mv "$TMPFILE" "${INSTALL_DIR}/${BIN_NAME}"

echo "Done! Run '${BIN_NAME} --install' to set up as a daemon."
