[CmdletBinding()]
param(
    [string]$ReleaseName = (Get-Date -Format 'yyyyMMdd-HHmm'),
    [switch]$SkipServer,
    [switch]$SkipWebAdmin,
    [switch]$SkipWebFront
)

$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$workspaceRoot = Split-Path -Parent $projectRoot
$deployRoot = Join-Path $workspaceRoot 'deploy-package'
$serverRoot = Join-Path $projectRoot 'server'
$webAdminRoot = Join-Path $projectRoot 'web'
$webFrontRoot = Join-Path $projectRoot 'web-front'
$archivePath = Join-Path $workspaceRoot ("MSTN-intranet-deployment-$ReleaseName.zip")

function Invoke-CheckedCommand {
    param(
        [string]$Name,
        [scriptblock]$Command
    )

    Write-Host "==> $Name"
    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE."
    }
}

function Replace-DirectoryContents {
    param(
        [string]$Source,
        [string]$Destination
    )

    if (-not (Test-Path -LiteralPath $Source -PathType Container)) {
        throw "Build output does not exist: $Source"
    }

    if (Test-Path -LiteralPath $Destination) {
        Remove-Item -LiteralPath $Destination -Recurse -Force
    }
    New-Item -ItemType Directory -Path $Destination -Force | Out-Null
    Copy-Item -Path (Join-Path $Source '*') -Destination $Destination -Recurse -Force
}

foreach ($path in @($deployRoot, $serverRoot, $webAdminRoot, $webFrontRoot)) {
    if (-not (Test-Path -LiteralPath $path -PathType Container)) {
        throw "Required directory does not exist: $path"
    }
}

if (Test-Path -LiteralPath $archivePath) {
    throw "Release archive already exists: $archivePath"
}

if (-not $SkipServer) {
    Invoke-CheckedCommand 'Building Go server' {
        Push-Location $serverRoot
        try {
            go build -o (Join-Path $deployRoot 'server/server.exe') main.go
        } finally {
            Pop-Location
        }
    }
    Replace-DirectoryContents -Source (Join-Path $serverRoot 'resource') -Destination (Join-Path $deployRoot 'server/resource')
}

if (-not $SkipWebAdmin) {
    Invoke-CheckedCommand 'Building admin web application' {
        $previousViteBase = $env:VITE_BASE
        $env:VITE_BASE = '/admin/'
        Push-Location $webAdminRoot
        try {
            npm run build
        } finally {
            Pop-Location
            $env:VITE_BASE = $previousViteBase
        }
    }
    Replace-DirectoryContents -Source (Join-Path $webAdminRoot 'dist') -Destination (Join-Path $deployRoot 'web-admin')
}

if (-not $SkipWebFront) {
    Invoke-CheckedCommand 'Building public web application' {
        Push-Location $webFrontRoot
        try {
            npm run build
        } finally {
            Pop-Location
        }
    }
    Replace-DirectoryContents -Source (Join-Path $webFrontRoot 'dist') -Destination (Join-Path $deployRoot 'web-front')
}

Write-Host '==> Creating release archive'
Write-Host "Archive path: $archivePath"
Compress-Archive -LiteralPath $deployRoot -DestinationPath $archivePath -CompressionLevel Optimal

Add-Type -AssemblyName System.IO.Compression.FileSystem
$archive = [System.IO.Compression.ZipFile]::OpenRead($archivePath)
try {
    $archiveFiles = @($archive.Entries | Where-Object { -not [string]::IsNullOrEmpty($_.Name) })
    $sourceFiles = @(Get-ChildItem -LiteralPath $deployRoot -File -Recurse)
    if ($archiveFiles.Count -ne $sourceFiles.Count) {
        throw "Archive validation failed: expected $($sourceFiles.Count) files, found $($archiveFiles.Count)."
    }
} finally {
    $archive.Dispose()
}

$archiveHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash
Write-Host "Release archive: $archivePath"
Write-Host "Files: $($sourceFiles.Count)"
Write-Host "SHA256: $archiveHash"
