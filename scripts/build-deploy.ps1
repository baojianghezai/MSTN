[CmdletBinding()]
param(
    [string]$ReleaseName = (Get-Date -Format 'yyyyMMdd-HHmm'),
    [switch]$SkipServer,
    [switch]$SkipWebAdmin,
    [switch]$SkipWebFront
)

# 打包产物：后端 exe + 两个前端静态站点，压缩为一个 zip。
# 结构：
#   server.exe
#   admin/   （VITE_BASE=/admin/ 构建）
#   www/   （同域部署构建）
$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$workspaceRoot = Split-Path -Parent $projectRoot
$serverRoot = Join-Path $projectRoot 'server'
$webAdminRoot = Join-Path $projectRoot 'web'
$webFrontRoot = Join-Path $projectRoot 'web-front'
$stagingRoot = Join-Path $workspaceRoot ("build-deploy-staging-$ReleaseName")
$archivePath = Join-Path $workspaceRoot ("MSTN-deployment-$ReleaseName.zip")

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

function Copy-BuildOutput {
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

foreach ($path in @($serverRoot, $webAdminRoot, $webFrontRoot)) {
    if (-not (Test-Path -LiteralPath $path -PathType Container)) {
        throw "Required directory does not exist: $path"
    }
}

if (Test-Path -LiteralPath $archivePath) {
    throw "Release archive already exists: $archivePath"
}

# 干净的暂存目录，只放要打包的三样东西
if (Test-Path -LiteralPath $stagingRoot) {
    Remove-Item -LiteralPath $stagingRoot -Recurse -Force
}
New-Item -ItemType Directory -Path $stagingRoot -Force | Out-Null

try {
    if (-not $SkipServer) {
        Invoke-CheckedCommand 'Building Go server' {
            Push-Location $serverRoot
            try {
                go build -o (Join-Path $stagingRoot 'server.exe') main.go
            } finally {
                Pop-Location
            }
        }
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
        Copy-BuildOutput -Source (Join-Path $webAdminRoot 'admin') -Destination (Join-Path $stagingRoot 'admin')
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
        Copy-BuildOutput -Source (Join-Path $webFrontRoot 'www') -Destination (Join-Path $stagingRoot 'www')
    }

    Write-Host '==> Creating release archive'
    Write-Host "Archive path: $archivePath"
    Compress-Archive -Path (Join-Path $stagingRoot '*') -DestinationPath $archivePath -CompressionLevel Optimal

    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive = [System.IO.Compression.ZipFile]::OpenRead($archivePath)
    try {
        $archiveFiles = @($archive.Entries | Where-Object { -not [string]::IsNullOrEmpty($_.Name) })
        $sourceFiles = @(Get-ChildItem -LiteralPath $stagingRoot -File -Recurse)
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
} finally {
    if (Test-Path -LiteralPath $stagingRoot) {
        Remove-Item -LiteralPath $stagingRoot -Recurse -Force
    }
}