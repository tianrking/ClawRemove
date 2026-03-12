$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Dist = Join-Path $Root "dist"
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

$Version = $env:CLAWREMOVE_VERSION
if ([string]::IsNullOrWhiteSpace($Version)) {
    try {
        $Version = (git describe --tags --always --dirty 2>$null)
    } catch {
        $Version = "dev"
    }
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $Version = "dev"
    }
}

$Targets = @(
  @{ GOOS = "darwin";  GOARCH = "amd64"; Ext = "";     ArchiveExt = ".tar.gz" },
  @{ GOOS = "darwin";  GOARCH = "arm64"; Ext = "";     ArchiveExt = ".tar.gz" },
  @{ GOOS = "linux";   GOARCH = "amd64"; Ext = "";     ArchiveExt = ".tar.gz" },
  @{ GOOS = "linux";   GOARCH = "arm64"; Ext = "";     ArchiveExt = ".tar.gz" },
  @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe"; ArchiveExt = ".zip" },
  @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe"; ArchiveExt = ".zip" }
)

# Clear existing checksums
$CheckSumFile = Join-Path $Dist "sha256sums.txt"
New-Item -ItemType File -Force -Path $CheckSumFile | Out-Null

Set-Location $Dist

foreach ($Target in $Targets) {
  $env:GOOS = $Target.GOOS
  $env:GOARCH = $Target.GOARCH
  $env:CGO_ENABLED = "0"
  
  $BinaryName = "claw-remove-{0}-{1}{2}" -f $Target.GOOS, $Target.GOARCH, $Target.Ext
  $Out = Join-Path $Dist $BinaryName
  
  Write-Host "==> Building $($Target.GOOS)/$($Target.GOARCH) (Version: $Version)"
  
  Set-Location $Root
  go build -trimpath -ldflags="-s -w -X github.com/tianrking/ClawRemove/internal/app.Version=$Version" -o $Out ./cmd/claw-remove
  
  Set-Location $Dist
  Write-Host "==> Packaging $($Target.GOOS)/$($Target.GOARCH)"
  $ArchiveName = "claw-remove-{0}-{1}{2}" -f $Target.GOOS, $Target.GOARCH, $Target.ArchiveExt
  
  if ($Target.GOOS -eq "windows") {
      Compress-Archive -Path $BinaryName -DestinationPath $ArchiveName -Force
  } else {
      # Fallback to tar if available on modern Windows
      tar -czf $ArchiveName $BinaryName
  }
  
  # Checksum
  $Hash = (Get-FileHash -Algorithm SHA256 -Path $ArchiveName).Hash.ToLower()
  $HashLine = "{0}  {1}" -f $Hash, $ArchiveName
  Add-Content -Path $CheckSumFile -Value $HashLine
}

Set-Location $Root
Write-Host "==> Release artifacts generated in $Dist"
