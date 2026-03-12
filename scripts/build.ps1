$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Dist = Join-Path $Root "dist"
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

$Targets = @(
  @{ GOOS = "darwin";  GOARCH = "amd64"; Ext = "" },
  @{ GOOS = "darwin";  GOARCH = "arm64"; Ext = "" },
  @{ GOOS = "linux";   GOARCH = "amd64"; Ext = "" },
  @{ GOOS = "linux";   GOARCH = "arm64"; Ext = "" },
  @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
  @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" }
)

foreach ($Target in $Targets) {
  $env:GOOS = $Target.GOOS
  $env:GOARCH = $Target.GOARCH
  $env:CGO_ENABLED = "0"
  $Out = Join-Path $Dist ("claw-remove-{0}-{1}{2}" -f $Target.GOOS, $Target.GOARCH, $Target.Ext)
  Write-Host "==> $($Target.GOOS)/$($Target.GOARCH)"
  go build -trimpath -ldflags="-s -w" -o $Out ./cmd/claw-remove
}
