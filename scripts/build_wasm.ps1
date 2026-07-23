# Build script for WASM deployment (Windows PowerShell)
# Usage: pwsh scripts/build_wasm.ps1

$ErrorActionPreference = "Stop"

Write-Host "Building WASM..." -ForegroundColor Cyan

$wasmdata = "cmd/termcom_wasm/wasmdata"
if (Test-Path $wasmdata) {
	Remove-Item -Recurse -Force $wasmdata
}
New-Item -ItemType Directory -Path $wasmdata -Force | Out-Null
Copy-Item -Recurse data $wasmdata\
Copy-Item -Recurse maps $wasmdata\

$env:GOOS = "js"
$env:GOARCH = "wasm"

$version = Get-Content VERSION -Raw
$version = $version.Trim()
$ldflags = "-ldflags=-X github.com/taislin/termcom/internal/engine.GameVersion=$version"

go build $ldflags -o web_wasm/termcom.wasm ./cmd/termcom_wasm/

Remove-Item -Recurse -Force $wasmdata

$goroot = go env GOROOT

$paths = @(
	"$goroot\lib\wasm\wasm_exec.js",
	"$goroot\misc\wasm\wasm_exec.js"
)

$found = $false
foreach ($p in $paths) {
	if (Test-Path $p) {
		Copy-Item $p web_wasm\wasm_exec.js
		$found = $true
		break
	}
}

if (-not $found) {
	Write-Host "WARNING: wasm_exec.js not found at expected paths" -ForegroundColor Yellow
}

$size = [math]::Round((Get-Item web_wasm/termcom.wasm).Length / 1MB, 1)
Write-Host ""
Write-Host "WASM build complete: web_wasm/" -ForegroundColor Green
Write-Host "  termcom.wasm   ($size MB)"
Write-Host "  wasm_exec.js"
Write-Host "  index.html"
Write-Host ""
Write-Host "To test locally: cd web_wasm; python3 -m http.server 8080"
