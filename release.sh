#!/bin/bash
# release.sh - Build and package MyStreamBot release
# Usage: ./release.sh <version>
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

BUILD_DIR="build/v$VERSION"
BINARY_NAME="mystreambot"

case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) BINARY_NAME="${BINARY_NAME}.exe" ;;
esac

echo "==> Building MyStreamBot v$VERSION..."

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "none")

go build -ldflags "-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE -X main.CommitHash=$COMMIT_HASH" -o "$BUILD_DIR/$BINARY_NAME" .

echo "==> Copying files..."

#cp  init.txt            "$BUILD_DIR/"
cp  twitchsubtypes.json "$BUILD_DIR/"
cp -r definitions       "$BUILD_DIR/"
cp -r modules           "$BUILD_DIR/"
cp -r web               "$BUILD_DIR/"
#cp -r db                "$BUILD_DIR/"
#cp -r logs              "$BUILD_DIR/"

echo "==> Done → $BUILD_DIR"
