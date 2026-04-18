param(
  [string]$Version = "",
  [string]$Repo = "Senasphy/ghstat-go",
  [string]$BinName = "ghstat-go",
  [string]$InstallDir = "$env:LOCALAPPDATA\Programs\ghstat-go\bin"
)

$ErrorActionPreference = "Stop"

function Resolve-Arch {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
  switch ($arch) {
    "x64" { return "amd64" }
    "arm64" { return "arm64" }
    default { throw "Unsupported architecture: $arch" }
  }
}

function Resolve-LatestVersion {
  $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
  return $latest.tag_name
}

if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = $env:GHSTAT_VERSION
}
if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = Resolve-LatestVersion
}
if ([string]::IsNullOrWhiteSpace($Version)) {
  throw "Failed to resolve release version"
}

$arch = Resolve-Arch
$os = "windows"
$artifact = "${BinName}_${Version}_${os}_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$Version/$artifact"

$tmpDir = Join-Path $env:TEMP ("ghstat-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
  $archivePath = Join-Path $tmpDir $artifact
  Invoke-WebRequest -Uri $url -OutFile $archivePath
  Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

  $sourceExe = Join-Path $tmpDir "$BinName.exe"
  if (-not (Test-Path $sourceExe)) {
    throw "Binary not found in archive: $sourceExe"
  }

  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  $targetExe = Join-Path $InstallDir "$BinName.exe"
  Copy-Item -Force $sourceExe $targetExe

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $pathEntries = @()
  if (-not [string]::IsNullOrWhiteSpace($userPath)) {
    $pathEntries = $userPath.Split(';') | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
  }

  if ($pathEntries -notcontains $InstallDir) {
    $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Added $InstallDir to User PATH. Restart your shell to apply it."
  }

  Write-Host "Installed $BinName $Version to $targetExe"
}
finally {
  if (Test-Path $tmpDir) {
    Remove-Item -Recurse -Force $tmpDir
  }
}
