#!/bin/bash
# fetch-cloudflared.sh — Download or build the QudsLab/Cloudflared shared library
#
# Usage:
#   ./fetch-cloudflared.sh [--arch=amd64|arm64] [--build]
#
# Options:
#   --arch=<arch>  Target architecture: amd64 or arm64 (default: host arch)
#   --build        Build from source instead of downloading pre-built binary
#
# The resulting library is placed at kernel/lib/libcloudflared.so (Linux),
# kernel/lib/libcloudflared.dylib (macOS), or kernel/lib/cloudflared.dll (Windows).
# Use -tags cflared when building the kernel to link against it.

set -e
trap 'echo "Error at line $LINENO: $BASH_COMMAND"; exit 1' ERR

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
KERNEL_DIR="$PROJECT_ROOT/kernel"
LIB_DIR="$KERNEL_DIR/lib"

ARCH="${GOARCH:-$(uname -m)}"
BUILD_FROM_SOURCE=false

# Normalise arch
case "$ARCH" in
    x86_64|amd64) ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

while [[ $# -gt 0 ]]; do
    case $1 in
        --arch=*) ARCH="${1#*=}"; shift ;;
        --build)  BUILD_FROM_SOURCE=true; shift ;;
        *) shift ;;
    esac
done

mkdir -p "$LIB_DIR"

# Map to QudsLab binary names
case "$OS" in
    linux)
        LIB_FILE="libcloudflared.so"
        QUDSLAB_OS="linux"
        ;;
    darwin)
        LIB_FILE="libcloudflared.dylib"
        QUDSLAB_OS="darwin"
        ;;
    mingw*|msys*|cygwin*|windows*)
        LIB_FILE="cloudflared.dll"
        QUDSLAB_OS="windows"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

DEST="$LIB_DIR/$LIB_FILE"

if [[ "$BUILD_FROM_SOURCE" == "true" ]]; then
    echo "Building cloudflared from source (QudsLab patch method)..."
    WORK_DIR="$(mktemp -d)"
    trap "rm -rf '$WORK_DIR'" EXIT

    echo "Cloning cloudflare/cloudflared..."
    git clone --depth 1 https://github.com/cloudflare/cloudflared.git "$WORK_DIR/cloudflared-src"

    echo "Cloning QudsLab/Cloudflared patches..."
    git clone --depth 1 https://github.com/QudsLab/Cloudflared.git "$WORK_DIR/quds"

    echo "Applying patches..."
    python3 "$WORK_DIR/quds/updates/replace.py" "$WORK_DIR/cloudflared-src"

    echo "Building c-shared library..."
    cd "$WORK_DIR/cloudflared-src"
    CGO_ENABLED=1 GOOS="$OS" GOARCH="$ARCH" \
        go build -buildmode=c-shared \
            -o "$DEST" \
            ./cmd/cloudflared

    echo "Copying header..."
    cp "$LIB_DIR/libcloudflared.h" "$LIB_DIR/" 2>/dev/null || true
    echo "Built: $DEST"
else
    # Download pre-built binary from QudsLab/Cloudflared
    echo "Downloading pre-built cloudflared library [$QUDSLAB_OS/$ARCH]..."
    BASE_URL="https://github.com/QudsLab/Cloudflared/raw/main/binaries/$QUDSLAB_OS/$ARCH"

    case "$OS" in
        linux)   REMOTE_FILE="libcloudflared.so" ;;
        darwin)  REMOTE_FILE="libcloudflared.dylib" ;;
        windows) REMOTE_FILE="cloudflared.dll" ;;
    esac

    curl -fsSL "$BASE_URL/$REMOTE_FILE" -o "$DEST"
    echo "Downloaded: $DEST"
fi

echo ""
echo "To build the kernel with Cloudflare tunnel support:"
echo "  CGO_ENABLED=1 go build --tags 'fts5 cflared' -ldflags \"-L$LIB_DIR\" ."
echo ""
echo "Ensure $LIB_FILE is in your library path at runtime:"
echo "  Linux:  export LD_LIBRARY_PATH=$LIB_DIR:\$LD_LIBRARY_PATH"
echo "  macOS:  export DYLD_LIBRARY_PATH=$LIB_DIR:\$DYLD_LIBRARY_PATH"
