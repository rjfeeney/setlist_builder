set -e
mkdir -p output
docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 -f Dockerfile.build -t setlist-builder-builder .
docker run --rm -v "$PWD/output":/app/output setlist-builder-builder
echo "Build complete. Binary is in ./output/"
