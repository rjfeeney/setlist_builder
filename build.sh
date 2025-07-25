set -e
mkdir -p output
docker build -f Dockerfile.build -t setlist-builder-builder .
docker run --rm -v "$PWD/output":/app/output setlist-builder-builder
echo "Build complete. Binary is in ./output/"
