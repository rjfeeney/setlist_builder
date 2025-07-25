#!/bin/bash
set -e
outputPath="./output"
binaryName="setlist"
imageName="setlist_builder_build_temp"
mkdir -p "$outputPath"
docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 -f Dockerfile.build -t $imageName .
containerId=$(docker create $imageName)
docker cp "$containerId:/app/setlist_builder" "$outputPath/$binaryName"
docker rm "$containerId"
echo "✅ Build complete! The binary is at $outputPath/$binaryName"
