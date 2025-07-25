$outputPath = "output"
if (-not (Test-Path $outputPath)) {
    New-Item -ItemType Directory -Path $outputPath | Out-Null
}
docker build --file Dockerfile.build `
             --build-arg GOOS=windows `
             --build-arg GOARCH=amd64 `
             -t setlist_builder_build_temp .
$containerId = docker create setlist_builder_build_temp
docker cp "$containerId:/app/setlist_builder" "$outputPath/setlist.exe"
docker rm $containerId | Out-Null
