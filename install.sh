#!/bin/sh
set -eu

REPO="ActionScripted/lazyopencode"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
case "$(uname -s)" in
  Linux*)  OS="linux"  ;;
  Darwin*) OS="darwin" ;;
  *)
           echo "error: unsupported OS: $(uname -s)"
                                                      exit 1
                                                             ;;
esac

# Detect architecture
case "$(uname -m)" in
  x86_64)  ARCH="amd64" ;;
  amd64)   ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
           echo "error: unsupported architecture: $(uname -m)"
                                                                exit 1
                                                                       ;;
esac

# Resolve version (latest release unless VERSION is set)
if [ -z "${VERSION:-}" ]; then
    VERSION=$(curl -fsSL -o /dev/null -w '%{url_effective}' \
        "https://github.com/${REPO}/releases/latest" | rev | cut -d/ -f1 | rev)
fi

BINARY="lazyopencode-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

echo "Installing lazyopencode ${VERSION} (${OS}/${ARCH})..."

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

curl -fsSL -o "$TMP" "$URL"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "${INSTALL_DIR}/lazyopencode"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$TMP" "${INSTALL_DIR}/lazyopencode"
fi

echo "Installed lazyopencode to ${INSTALL_DIR}/lazyopencode"
