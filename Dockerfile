$outputPath = "output"
if (-not (Test-Path $outputPath)) {
    New-Item -ItemType Directory -Path $outputPath | Out-Null
}
docker build -t setlist_builder_build -f Dockerfile.build .
docker run --rm -v "${PWD}\$outputPath:/app/output" setlist_builder_build go build -o /app/output/setlist.exe
