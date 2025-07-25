$outputPath = "output"
if (-not (Test-Path -Path $outputPath)) {
    New-Item -ItemType Directory -Path $outputPath | Out-Null
}
docker build -t setlist_builder_build_temp -f Dockerfile.build .
$containerId = docker create setlist_builder_build_temp
docker cp "$containerId:/app/output/setlist" "$outputPath/setlist.exe"
docker rm $containerId
