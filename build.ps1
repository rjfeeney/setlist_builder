if (!(Test-Path -Path "output")) {
    New-Item -ItemType Directory -Path "output"
}
docker build -f Dockerfile.build -t setlist-builder-builder .
docker run --rm -v ${PWD}\output:/app/output setlist-builder-builder
Write-Host "Build complete. Binary is in ./output/"
