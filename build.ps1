Write-Host "🛠 Building setlist CLI for Windows..."
$outputPath = "output"
if (-not (Test-Path $outputPath)) {
    New-Item -ItemType Directory -Path $outputPath | Out-Null
}
docker build --file Dockerfile.build `
             --build-arg GOOS=windows `
             --build-arg GOARCH=amd64 `
             -t setlist_builder_build_temp .
docker run --rm -v "$(Join-Path $PWD 'output'):/out" setlist_builder_build_temp
Write-Host "✅ Build complete. Windows binary is in output\setlist.exe"
