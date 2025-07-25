#!/bin/bash
set -e
outputPath="./output"
mkdir -p "$outputPath"
docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 -f Dockerfile.build -t setlist_builder_build_temp .
containerId=$(docker create setlist_builder_build_temp)
docker cp "$containerId:/app/output/setlist_builder" "$fullOutputPath\setlist.exe"
docker rm "$containerId"
echo "Build complete! The binary is at $outputPath/setlist"

