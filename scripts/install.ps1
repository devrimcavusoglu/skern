# Skern installer - Windows / PowerShell counterpart to scripts/install.sh
# Usage:
#   irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex
#
# Environment variables:
#   SKERN_INSTALL_DIR  - override install directory (default: %LOCALAPPDATA%\skern\bin)
#   SKERN_VERSION      - install a specific version (default: latest)

$ErrorActionPreference = 'Stop'

$Repo = 'devrimcavusoglu/skern'
$Binary = 'skern'
$DefaultInstallDir = Join-Path $env:LOCALAPPDATA 'skern\bin'

# PowerShell 5.1 default protocol may be TLS 1.0/1.1; GitHub requires 1.2+.
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

function Write-Info($msg) {
    Write-Host "[skern] $msg"
}

function Stop-WithError($msg) {
    Write-Host "[skern] ERROR: $msg" -ForegroundColor Red
    exit 1
}

function Get-Arch {
    $procArch = $env:PROCESSOR_ARCHITECTURE
    $procArchW6432 = $env:PROCESSOR_ARCHITEW6432
    if ($procArch -eq 'ARM64' -or $procArchW6432 -eq 'ARM64') {
        return 'arm64'
    }
    if ([System.Environment]::Is64BitOperatingSystem) {
        return 'amd64'
    }
    Stop-WithError "Unsupported architecture: $procArch. Skern supports amd64 and arm64."
}

function Resolve-Version {
    if ($env:SKERN_VERSION) {
        Write-Info "Using specified version: $($env:SKERN_VERSION)"
        return $env:SKERN_VERSION
    }
    Write-Info 'Fetching latest release...'
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    } catch {
        Stop-WithError "Could not fetch latest release: $_"
    }
    if (-not $release.tag_name) {
        Stop-WithError "Could not determine latest version. Set SKERN_VERSION manually or use 'go install'."
    }
    Write-Info "Latest version: $($release.tag_name)"
    return $release.tag_name
}

function Get-AssetWithChecksum($version, $arch, $tmpDir) {
    $versionNoV = $version -replace '^v',''
    $assetName = "skern_${versionNoV}_windows_${arch}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$assetName"
    $checksumsUrl = "https://github.com/$Repo/releases/download/$version/checksums.txt"

    $zipPath = Join-Path $tmpDir $assetName
    $checksumsPath = Join-Path $tmpDir 'checksums.txt'

    Write-Info "Downloading $downloadUrl ..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
    } catch {
        Stop-WithError "Download failed: $_"
    }

    try {
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
    } catch {
        Write-Info 'Warning: checksums file not available, skipping verification.'
        $checksumsPath = $null
    }

    if ($checksumsPath -and (Test-Path $checksumsPath)) {
        $expectedLine = Get-Content $checksumsPath | Where-Object { $_ -match [regex]::Escape($assetName) } | Select-Object -First 1
        if ($expectedLine) {
            $expectedSum = ($expectedLine -split '\s+')[0]
            $actualSum = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
            if ($expectedSum.ToLower() -ne $actualSum) {
                Stop-WithError "Checksum mismatch! Expected $expectedSum, got $actualSum. Aborting."
            }
            Write-Info 'Checksum verified.'
        } else {
            Write-Info "Warning: could not find checksum for $assetName, skipping verification."
        }
    }

    return $zipPath
}

function Install-Binary($zipPath, $tmpDir) {
    $installDir = if ($env:SKERN_INSTALL_DIR) { $env:SKERN_INSTALL_DIR } else { $DefaultInstallDir }
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    $extractDir = Join-Path $tmpDir 'extracted'
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $exeName = "$Binary.exe"
    $sourceExe = Get-ChildItem -Path $extractDir -Filter $exeName -Recurse | Select-Object -First 1
    if (-not $sourceExe) {
        Stop-WithError "Binary $exeName not found after extraction. Archive may be corrupt."
    }

    $targetPath = Join-Path $installDir $exeName
    Move-Item -Path $sourceExe.FullName -Destination $targetPath -Force
    Write-Info "Installed $exeName to $targetPath"

    # Check if installDir is on PATH (user or system).
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $machinePath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
    $combinedPath = "$userPath;$machinePath"
    $onPath = $combinedPath.Split(';') | Where-Object { $_.TrimEnd('\') -ieq $installDir.TrimEnd('\') }

    if (-not $onPath) {
        Write-Info ''
        Write-Info "WARNING: $installDir is not on your PATH."
        Write-Info 'Add it persistently with:'
        Write-Info ''
        Write-Info "  [Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$installDir`", 'User')"
        Write-Info ''
        Write-Info 'Then open a new shell. Or, for the current shell only:'
        Write-Info ''
        Write-Info "  `$env:Path += ';$installDir'"
        Write-Info ''
    }

    return $targetPath
}

function Main {
    Write-Info 'Skern installer'
    Write-Info ''

    $arch = Get-Arch
    Write-Info "Detected platform: windows_$arch"

    $version = Resolve-Version

    $tmpDir = Join-Path $env:TEMP "skern-install-$([System.Guid]::NewGuid().ToString('N'))"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        $zipPath = Get-AssetWithChecksum -version $version -arch $arch -tmpDir $tmpDir
        $targetPath = Install-Binary -zipPath $zipPath -tmpDir $tmpDir

        Write-Info ''
        try {
            $output = & $targetPath version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Info "Success! Installed $output"
            } else {
                Write-Info 'Installation complete. You may need to open a new shell to use skern.'
            }
        } catch {
            Write-Info 'Installation complete. You may need to open a new shell to use skern.'
        }
    } finally {
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

Main
