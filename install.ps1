<#
.SYNOPSIS
    Installs asbuilt for Windows from a GitHub release.

.DESCRIPTION
    Downloads asbuilt-<version>-windows-amd64.exe from the latest (or a given)
    GitHub release, verifies it against the release checksums.txt, installs it
    into %LOCALAPPDATA%\Programs\ASBUILT and adds that directory to the user
    PATH when it is not there yet.

    The script needs no administrator rights and can be run again at any time:
    it replaces the binary in place.

.EXAMPLE
    irm https://raw.githubusercontent.com/Balanov/asbuilt/main/install.ps1 | iex

.EXAMPLE
    .\install.ps1 -Version 0.1.0
#>
param(
    [string]$Version = ""
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 2.0

$Repo = 'Balanov/asbuilt'
$ApiBase = "https://api.github.com/repos/$Repo"
$DownloadBase = "https://github.com/$Repo/releases/download"
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\ASBUILT'
$Target = 'windows-amd64'

function Write-Info {
    param([string]$Message)
    Write-Host "asbuilt-install: $Message"
}

function Write-Warn {
    param([string]$Message)
    Write-Host "asbuilt-install: warning: $Message" -ForegroundColor Yellow
}

function Get-LatestTag {
    $headers = @{
        'User-Agent' = 'asbuilt-install'
        'Accept'     = 'application/vnd.github+json'
    }
    $release = Invoke-RestMethod -Uri "$ApiBase/releases/latest" -Headers $headers -UseBasicParsing
    if (-not $release.tag_name) {
        throw "the latest release carries no tag_name"
    }
    return [string]$release.tag_name
}

function Get-Sha256 {
    param([string]$Path)
    return (Get-FileHash -Path $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Test-Checksum {
    param(
        [string]$Tag,
        [string]$Asset,
        [string]$File,
        [string]$TempDir
    )
    $sumsPath = Join-Path $TempDir 'checksums.txt'
    try {
        Invoke-WebRequest -Uri "$DownloadBase/$Tag/checksums.txt" -OutFile $sumsPath -UseBasicParsing
    } catch {
        Write-Warn "could not download checksums.txt ($($_.Exception.Message)), skipping checksum verification"
        return
    }

    $expected = $null
    foreach ($line in (Get-Content -Path $sumsPath)) {
        $parts = $line.Trim() -split '\s+'
        if ($parts.Count -ge 2) {
            $name = $parts[1] -replace '^\*', ''
            if ($name -eq $Asset) {
                $expected = $parts[0].ToLowerInvariant()
                break
            }
        }
    }
    if (-not $expected) {
        Write-Warn "no checksum is published for $Asset, skipping checksum verification"
        return
    }

    $actual = Get-Sha256 -Path $File
    if ($actual -ne $expected) {
        Write-Warn "checksum mismatch for $Asset"
        Write-Warn "  expected: $expected"
        Write-Warn "  actual:   $actual"
        return
    }
    Write-Info "checksum verified"
}

function Add-ToUserPath {
    param([string]$Directory)
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($null -eq $userPath) {
        $userPath = ''
    }
    $entries = @($userPath -split ';' | Where-Object { $_ -ne '' })
    $wanted = $Directory.TrimEnd('\', '/')
    foreach ($entry in $entries) {
        if ($entry.TrimEnd('\', '/') -ieq $wanted) {
            Write-Info "$Directory is already in the user PATH"
            return
        }
    }
    $newPath = if ($userPath) { "$userPath;$Directory" } else { $Directory }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Info "added $Directory to the user PATH"
    Write-Warn "open a new terminal for the PATH change to take effect"
}

$tag = $Version
if ([string]::IsNullOrWhiteSpace($tag)) {
    $tag = Get-LatestTag
}
if ($tag.StartsWith('v')) {
    $version = $tag.Substring(1)
} else {
    $version = $tag
    $tag = "v$version"
}

$asset = "asbuilt-$version-$Target.exe"
$url = "$DownloadBase/$tag/$asset"

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("asbuilt-install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    Write-Info "installing asbuilt $version for $Target"
    Write-Info "downloading $url"
    $downloaded = Join-Path $tempDir $asset
    Invoke-WebRequest -Uri $url -OutFile $downloaded -UseBasicParsing

    Test-Checksum -Tag $tag -Asset $asset -File $downloaded -TempDir $tempDir

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $downloaded -Destination (Join-Path $InstallDir $asset) -Force
    $exe = Join-Path $InstallDir 'asbuilt.exe'
    Copy-Item -Path $downloaded -Destination $exe -Force

    Add-ToUserPath -Directory $InstallDir

    Write-Info "installed asbuilt $version to $exe"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  1. Open a new terminal (the PATH change needs a new session)."
    Write-Host "  2. asbuilt --version"
    Write-Host "  3. asbuilt check"
    Write-Host "  4. asbuilt rules"
} finally {
    if (Test-Path -Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
