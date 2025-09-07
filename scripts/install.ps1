# Usage (1-liner):
#   powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"

param(
  [string]$Owner = 'Hangell',
  [string]$Repo  = 'gommit',
  [switch]$PreRelease,
  [switch]$Debug
)

$ErrorActionPreference = 'Continue'  # Changed to Continue for better debugging
$ua = @{ 'User-Agent' = 'gommit-installer' }

function Write-DebugInfo {
  param([string]$Message, [string]$Color = 'Cyan')
  Write-Host "[DEBUG] $Message" -ForegroundColor $Color
}

function Write-Step {
  param([string]$Message)
  Write-Host "🔄 $Message" -ForegroundColor Yellow
}

function Write-Success {
  param([string]$Message)
  Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Failure {
  param([string]$Message)
  Write-Host "❌ $Message" -ForegroundColor Red
}

try {
  Write-Success "Starting gommit installation..."

  # Detecta SO/arch
  Write-Step "Detecting system architecture..."
  $os = 'windows'
  
  Write-DebugInfo "Calling RuntimeInformation..."
  $archInfo = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
  Write-DebugInfo "Got architecture info: $archInfo"
  
  $arch = $archInfo.ToString().ToLower()
  Write-DebugInfo "Architecture string: $arch"
  
  switch ($arch) {
    'x64'   { $arch = 'amd64' }
    'arm64' { $arch = 'arm64' }
    default { $arch = 'amd64' }
  }

  Write-Success "Detected: ${os}_${arch}"

  # Pega release
  Write-Step "Fetching release information..."
  $api = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
  if ($PreRelease) { 
    $api = "https://api.github.com/repos/$Owner/$Repo/releases" 
    Write-DebugInfo "Using pre-release endpoint"
  }

  Write-DebugInfo "API URL: $api"
  
  Write-DebugInfo "Making API request..."
  $resp = $null
  try {
    $resp = Invoke-RestMethod -UseBasicParsing -Headers $ua -Uri $api -TimeoutSec 30
    Write-Success "API request successful"
  } catch {
    Write-Failure "API request failed: $($_.Exception.Message)"
    Write-DebugInfo "Exception type: $($_.Exception.GetType().Name)"
    Write-DebugInfo "Status code: $($_.Exception.Response.StatusCode)"
    return
  }

  Write-DebugInfo "Response type: $($resp.GetType().Name)"
  
  if ($PreRelease) { 
    Write-Step "Looking for pre-release..."
    $preReleases = $resp | Where-Object { $_.prerelease -eq $true }
    Write-DebugInfo "Found $($preReleases.Count) pre-releases"
    $resp = $preReleases | Select-Object -First 1
    if (-not $resp) {
      Write-Failure "No pre-release found"
      return
    }
  }

  Write-DebugInfo "Release tag: $($resp.tag_name)"
  Write-DebugInfo "Release name: $($resp.name)"

  # Debug assets
  Write-Step "Analyzing release assets..."
  if (-not $resp.assets) {
    Write-Failure "No assets property found in response"
    Write-DebugInfo "Response keys: $($resp | Get-Member -MemberType Properties | Select-Object -ExpandProperty Name)"
    return
  }

  Write-DebugInfo "Assets type: $($resp.assets.GetType().Name)"
  Write-DebugInfo "Assets count: $($resp.assets.Count)"
  
  if ($resp.assets.Count -eq 0) {
    Write-Failure "No assets found in the release"
    return
  }

  Write-Success "Found $($resp.assets.Count) assets:"
  foreach ($asset in $resp.assets) {
    Write-Host "  📦 $($asset.name)" -ForegroundColor Cyan
  }

  # Escolhe asset
  Write-Step "Looking for matching asset..."
  $pattern = "gommit_.*_${os}_${arch}\.zip$"
  Write-DebugInfo "Search pattern: $pattern"
  
  $matchingAssets = @()
  foreach ($asset in $resp.assets) {
    if ($asset.name -match $pattern) {
      $matchingAssets += $asset
      Write-DebugInfo "✅ MATCH: $($asset.name)"
    } else {
      Write-DebugInfo "❌ NO MATCH: $($asset.name)"
    }
  }

  if ($matchingAssets.Count -eq 0) {
    Write-Failure "No matching asset found for pattern: $pattern"
    Write-Host "Available assets:" -ForegroundColor Red
    foreach ($asset in $resp.assets) {
      Write-Host "  - $($asset.name)" -ForegroundColor Red
    }
    return
  }

  $selectedAsset = $matchingAssets[0]
  Write-Success "Selected asset: $($selectedAsset.name)"

  # Download
  Write-Step "Preparing download..."
  $tmp = Join-Path $env:TEMP ("gommit-" + [guid]::NewGuid())
  Write-DebugInfo "Temp directory: $tmp"
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  $zip = Join-Path $tmp "gommit.zip"

  Write-Step "Downloading $($selectedAsset.name)..."
  Write-DebugInfo "Download URL: $($selectedAsset.browser_download_url)"
  
  try {
    Invoke-WebRequest -UseBasicParsing -Uri $selectedAsset.browser_download_url -OutFile $zip -TimeoutSec 60
    Write-Success "Download completed"
    Write-DebugInfo "File size: $((Get-Item $zip).Length) bytes"
  } catch {
    Write-Failure "Download failed: $($_.Exception.Message)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    return
  }

  # Extract
  Write-Step "Extracting archive..."
  try {
    Expand-Archive -Path $zip -DestinationPath $tmp -Force
    Write-Success "Extraction completed"
  } catch {
    Write-Failure "Extraction failed: $($_.Exception.Message)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    return
  }

  # Find executable
  Write-Step "Looking for gommit.exe..."
  Write-DebugInfo "Searching in: $tmp"
  
  $allFiles = Get-ChildItem -Path $tmp -Recurse
  Write-DebugInfo "All files found:"
  foreach ($file in $allFiles) {
    Write-Host "  📄 $($file.FullName)" -ForegroundColor Gray
  }

  $exe = Get-ChildItem -Path $tmp -Recurse -Filter gommit.exe | Select-Object -First 1
  
  if (-not $exe) { 
    Write-Failure "gommit.exe not found in extracted files"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    return
  }

  Write-Success "Found executable: $($exe.FullName)"

  # Install
  Write-Step "Running installation..."
  try {
    Write-DebugInfo "Executing: $($exe.FullName) --install"
    & $exe.FullName --install
    $installExitCode = $LASTEXITCODE
    Write-DebugInfo "Install exit code: $installExitCode"
    
    if ($installExitCode -ne 0) {
      Write-Failure "Installation failed with exit code: $installExitCode"
      Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
      return
    }
    Write-Success "Installation completed successfully"
  } catch {
    Write-Failure "Failed to run installation: $($_.Exception.Message)"
    Write-DebugInfo "Exception type: $($_.Exception.GetType().Name)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    return
  }

  # Cleanup
  Write-Step "Cleaning up..."
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue

  Write-Host ""
  Write-Success "Installation completed successfully!"
  Write-Host "Close and reopen the terminal, then run: gommit --version" -ForegroundColor Yellow

} catch {
  Write-Failure "Unexpected error occurred"
  Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
  Write-Host "Exception type: $($_.Exception.GetType().Name)" -ForegroundColor Red
  Write-Host "Line: $($_.InvocationInfo.ScriptLineNumber)" -ForegroundColor Red
  Write-Host "Position: $($_.InvocationInfo.PositionMessage)" -ForegroundColor Red
  if ($_.ScriptStackTrace) {
    Write-Host "Stack trace:" -ForegroundColor Red
    Write-Host $_.ScriptStackTrace -ForegroundColor Red
  }
}