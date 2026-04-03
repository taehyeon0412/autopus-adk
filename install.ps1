# autopus-adk Windows install script
# Usage: irm https://raw.githubusercontent.com/Insajin/autopus-adk/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "Insajin/autopus-adk"
$Binary = "auto.exe"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\autopus-adk\bin" }

function Info($msg)  { Write-Host $msg -ForegroundColor Cyan }
function Ok($msg)    { Write-Host $msg -ForegroundColor Green }
function Err($msg)   { Write-Host $msg -ForegroundColor Red; exit 1 }

# Detect architecture
function Get-Arch {
    $envArch = $env:PROCESSOR_ARCHITECTURE
    switch ($envArch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Err "Unsupported architecture: $envArch" }
    }
}

# Get latest version from GitHub API
function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ "User-Agent" = "autopus-installer" }
    return $release.tag_name -replace '^v', ''
}

# Verify SHA256 checksum
function Verify-Checksum($file, $expected) {
    $actual = (Get-FileHash -Path $file -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $expected) {
        Err "Checksum mismatch!`n  expected: $expected`n  actual:   $actual"
    }
}

function Main {
    $Arch = Get-Arch
    $Version = if ($env:VERSION) { $env:VERSION } else { Get-LatestVersion }

    if (-not $Version) {
        Err "Failed to get latest version. Check GitHub API limits."
    }

    Info "autopus-adk v$Version installing... (windows/$Arch)"

    $Archive = "autopus-adk_${Version}_windows_${Arch}.zip"
    $BaseUrl = "https://github.com/$Repo/releases/download/v$Version"
    $Url = "$BaseUrl/$Archive"
    $ChecksumsUrl = "$BaseUrl/checksums.txt"

    $TmpDir = Join-Path $env:TEMP "autopus-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

    try {
        Info "Downloading: $Url"
        Invoke-WebRequest -Uri $Url -OutFile "$TmpDir\$Archive" -UseBasicParsing

        # SHA256 checksum verification
        Info "Verifying checksum..."
        Invoke-WebRequest -Uri $ChecksumsUrl -OutFile "$TmpDir\checksums.txt" -UseBasicParsing
        $checksumLine = Get-Content "$TmpDir\checksums.txt" | Where-Object { $_ -match $Archive }
        if ($checksumLine) {
            $expected = ($checksumLine -split '\s+')[0].ToLower()
            Verify-Checksum "$TmpDir\$Archive" $expected
            Ok "Checksum verified"
        } else {
            Err "Checksum not found for $Archive in checksums.txt"
        }

        Info "Extracting..."
        Expand-Archive -Path "$TmpDir\$Archive" -DestinationPath $TmpDir -Force

        # Create install directory if it doesn't exist
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        Info "Installing to $InstallDir\$Binary..."
        $TargetPath = "$InstallDir\$Binary"
        $OldPath = "$TargetPath.old"
        if (Test-Path $TargetPath) {
            # Running exe cannot be overwritten but CAN be renamed.
            Remove-Item $OldPath -Force -ErrorAction SilentlyContinue
            try {
                Rename-Item $TargetPath $OldPath -Force
            } catch {
                # Rename failed — try direct copy as last resort.
            }
        }
        Copy-Item "$TmpDir\auto.exe" $TargetPath -Force
        Remove-Item $OldPath -Force -ErrorAction SilentlyContinue

        # Add to PATH if not already present
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
            Info "Added $InstallDir to user PATH"
        }

        Ok "autopus-adk v$Version installed!"
        Ok ""

        # Post-install: check and install dependencies (skip already installed)
        Info "Checking dependencies..."
        try {
            & "$InstallDir\$Binary" doctor --fix --yes 2>$null
            Ok "Dependencies installed!"
        } catch {
            Write-Host "  Some dependencies could not be auto-installed." -ForegroundColor Yellow
            Write-Host "  Run manually: auto doctor" -ForegroundColor Yellow
        }

        # Auto-init: detect platform and initialize harness
        if ($env:SKIP_INIT -eq "1") {
            Ok ""
            Ok "  SKIP_INIT=1 — skipping initialization."
            Ok "  Next: auto init"
            Ok ""
        }
        elseif ((Test-Path "CLAUDE.md") -or (Test-Path "autopus.yaml")) {
            Ok "Already initialized. Running update..."
            try { & "$InstallDir\$Binary" update --yes 2>$null } catch {}
            Ok ""
            Ok "  Ready to use:"
            Ok "    /auto setup    # generate project context"
            Ok "    /auto status   # SPEC dashboard"
            Ok ""
        }
        else {
            Info "Initializing project..."
            $Proj = if ($env:PROJECT_NAME) { $env:PROJECT_NAME } else { Split-Path -Leaf (Get-Location) }
            $Plat = if ($env:PLATFORMS) { $env:PLATFORMS } else { "claude-code" }
            # Detect additional platforms
            if (-not $env:PLATFORMS) {
                if (Get-Command codex -ErrorAction SilentlyContinue) { $Plat += ",codex" }
                if (Get-Command gemini -ErrorAction SilentlyContinue) { $Plat += ",gemini" }
            }
            Info "  Project: $Proj"
            Info "  Platforms: $Plat"
            try {
                & "$InstallDir\$Binary" init --project $Proj --platforms $Plat --yes 2>&1
                Ok "Project initialized!"
            } catch {
                Write-Host "  Init failed. Run manually: auto init" -ForegroundColor Yellow
            }
            Ok ""
            Ok "  Ready to use in Claude Code:"
            Ok "    /auto setup    # generate project context"
            Ok "    /auto plan     # write a SPEC"
            Ok "    /auto fix      # fix a bug"
            Ok "    /auto review   # code review"
            Ok ""
        }
    }
    finally {
        Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Main
