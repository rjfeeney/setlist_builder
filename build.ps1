$outputDir = Join-Path -Path (Get-Location) -ChildPath "output"
$binaryName = "setlist.exe"
$imageName = "setlist_builder_build_temp"
$containerName = "setlist_builder_build_container"
$binaryPathInContainer = "/app/setlist_builder"

if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir | Out-Null
}

docker build -f Dockerfile.build -t $imageName .

$containerId = docker create --name $containerName $imageName

if (-not $containerId) {
    Write-Error "Failed to create Docker container."
    exit 1
}

docker cp "$containerName`:$binaryPathInContainer" (Join-Path $outputDir $binaryName)

docker rm $containerName | Out-Null

Write-Host "✅ Build complete! The binary is at $outputDir\$binaryName"
