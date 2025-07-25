$outputPath = ".\output"
if (-not (Test-Path $outputPath)) {
    New-Item -ItemType Directory -Path $outputPath | Out-Null
}
docker build --build-arg GOOS=windows --build-arg GOARCH=amd64 -f Dockerfile.build -t setlist_builder_build_temp .
$containerId = docker create setlist_builder_build_temp
$sourcePath = "/app/output/setlist_builder"
$destinationPath = "$outputPath\setlist.exe"
docker cp $containerId:$sourcePath $destinationPath
docker rm $containerId
Write-Host "Build complete! The binary is at $destinationPath"

