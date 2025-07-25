#!/bin/bash
set -e
mkdir -p output
UNAME=$(uname | tr '[:upper:]' '[:lower:]')
if [[ "$UNAME" == "darwin" ]]; then
  GOOS="darwin"
elif [[ "$UNAME" == "linux" ]]; then
  GOOS="linux"
else
  echo "Unsupported OS: $UNAME"
  exit 1
fi
GOARCH="amd64"
echo "Building for OS=$GOOS ARCH=$GOARCH"
docker build --build-arg GOOS=$GOOS --build-arg GOARCH=$GOARCH -f Dockerfile.build -t setlist-builder-builder .
docker run --rm -v "$PWD/output":/app/output setlist-builder-builder
echo "Build complete. Binary is in ./output/"
