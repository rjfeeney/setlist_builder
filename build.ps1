if (!(Test-Path -Path "output")) {
    New-Item -ItemType Directory -Path "output" | Out-Null
}
$platform = $PSVersionTable.Platform
switch ($platform) {
    "Win32NT" {
        $GOOS = "windows"
        $GOARCH = "amd64"
    }
    "Unix" {
        $GOOS = "linux"
        $GOARCH = "amd64"
    }
    default {
        Write-Error "Unsupported platform: $platform"
        exit 1
    }
}
Write-Host "Building for OS=$GOOS ARCH=$GOARCH"
docker build --build-arg GOOS=$GOOS --build-arg GOARCH=$GOARCH -f Dockerfile.build -t setlist-builder-builder .
docker run --rm -v ${PWD}\output:/app/output setlist-builder-builder
Write-Host "Build complete. Binary is in ./output/"
