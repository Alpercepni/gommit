# Usage (1-liner):
#   powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"

param(
  [string]$Owner = 'Hangell',
  [string]$Repo  = 'gommit',
  [switch]$PreRelease # usa o último pré-release
)

$ErrorActionPreference = 'Stop'
$ua = @{ 'User-Agent' = 'gommit-installer' }

try {
  Write-Host "Starting gommit installation..." -ForegroundColor Green

  # Detecta SO/arch
  $os = 'windows'
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLower()
  
  Write-Host "Detected architecture: $arch" -ForegroundColor Yellow
  
  switch ($arch) {
    'x64'   { $arch = 'amd64' }
    'arm64' { $arch = 'arm64' }
    default { $arch = 'amd64' }
  }

  Write-Host "Using architecture: $arch" -ForegroundColor Yellow

  # Pega release
  $api = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
  if ($PreRelease) { 
    $api = "https://api.github.com/repos/$Owner/$Repo/releases" 
    Write-Host "Looking for pre-release..." -ForegroundColor Yellow
  }

  Write-Host "Fetching release info from: $api" -ForegroundColor Yellow
  
  try {
    $resp = Invoke-RestMethod -UseBasicParsing -Headers $ua -Uri $api -TimeoutSec 30
  } catch {
    Write-Error "Failed to fetch release information: $($_.Exception.Message)"
    Write-Host "Please check if the repository exists and has releases: https://github.com/$Owner/$Repo/releases" -ForegroundColor Red
    exit 1
  }

  if ($PreRelease) { 
    $resp = ($resp | Where-Object { $_.prerelease -eq $true } | Select-Object -First 1)
    if (-not $resp) {
      Write-Error "No pre-release found"
      exit 1
    }
  }

  Write-Host "Found release: $($resp.tag_name)" -ForegroundColor Green

  # Escolhe asset .zip do OS/arch
  $pattern = "gommit_.*_${os}_${arch}\.zip$"
  Write-Host "Looking for asset matching pattern: $pattern" -ForegroundColor Yellow
  
  if (-not $resp.assets -or $resp.assets.Count -eq 0) {
    Write-Error "No assets found in the release"
    Write-Host "Available assets:" -ForegroundColor Red
    if ($resp.assets) {
      $resp.assets | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor Red }
    } else {
      Write-Host "  (none)" -ForegroundColor Red
    }
    exit 1
  }

  Write-Host "Available assets:" -ForegroundColor Yellow
  $resp.assets | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor Yellow }

  $asset = $resp.assets | Where-Object { $_.name -match $pattern } | Select-Object -First 1
  
  if (-not $asset) { 
    Write-Error "No matching asset for ${os}_${arch}"
    Write-Host "Looking for pattern: $pattern" -ForegroundColor Red
    Write-Host "Available assets:" -ForegroundColor Red
    $resp.assets | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor Red }
    exit 1
  }

  Write-Host "Selected asset: $($asset.name)" -ForegroundColor Green

  # Baixa e extrai
  $tmp = Join-Path $env:TEMP ("gommit-" + [guid]::NewGuid())
  Write-Host "Creating temp directory: $tmp" -ForegroundColor Yellow
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  $zip = Join-Path $tmp "gommit.zip"

  Write-Host "Downloading $($asset.name)..." -ForegroundColor Green
  try {
    Invoke-WebRequest -UseBasicParsing -Uri $asset.browser_download_url -OutFile $zip -TimeoutSec 60
    Write-Host "Download completed successfully" -ForegroundColor Green
  } catch {
    Write-Error "Failed to download: $($_.Exception.Message)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
  }

  Write-Host "Extracting archive..." -ForegroundColor Yellow
  try {
    Expand-Archive -Path $zip -DestinationPath $tmp -Force
    Write-Host "Extraction completed" -ForegroundColor Green
  } catch {
    Write-Error "Failed to extract archive: $($_.Exception.Message)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
  }

  # Procura o binário
  Write-Host "Looking for gommit.exe..." -ForegroundColor Yellow
  $exe = Get-ChildItem -Path $tmp -Recurse -Filter gommit.exe | Select-Object -First 1
  
  if (-not $exe) { 
    Write-Host "Contents of temp directory:" -ForegroundColor Red
    Get-ChildItem -Path $tmp -Recurse | ForEach-Object { Write-Host "  $($_.FullName)" -ForegroundColor Red }
    Write-Error "gommit.exe not found in package"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
  }

  Write-Host "Found executable: $($exe.FullName)" -ForegroundColor Green

  # Instala
  Write-Host "Installing gommit..." -ForegroundColor Green
  try {
    & $exe.FullName --install
    $installExitCode = $LASTEXITCODE
    if ($installExitCode -ne 0) {
      Write-Error "Installation failed with exit code: $installExitCode"
      Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
      exit 1
    }
    Write-Host "Installation completed successfully" -ForegroundColor Green
  } catch {
    Write-Error "Failed to run installation: $($_.Exception.Message)"
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
  }

  # Limpa
  Write-Host "Cleaning up..." -ForegroundColor Yellow
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue

  Write-Host ""
  Write-Host "✅ Installation completed successfully!" -ForegroundColor Green
  Write-Host "Close and reopen the terminal, then run: gommit --version" -ForegroundColor Yellow

} catch {
  Write-Error "Installation failed: $($_.Exception.Message)"
  Write-Host "Stack trace:" -ForegroundColor Red
  Write-Host $_.ScriptStackTrace -ForegroundColor Red
  exit 1
}