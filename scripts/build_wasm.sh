#!/usr/bin/env bash
# Build script for WASM deployment (cross-platform)
# Usage: ./scripts/build_wasm.sh

set -e
cd "$(dirname "$0")/.."

echo "Building WASM..."

WASMDATA="cmd/termcom_wasm/wasmdata"
rm -rf "$WASMDATA"
mkdir -p "$WASMDATA"
cp -r data "$WASMDATA/"
cp -r maps "$WASMDATA/"

GOOS=js GOARCH=wasm go build -ldflags="-X github.com/taislin/termcom/internal/engine.GameVersion=$(cat VERSION)" -o web_wasm/termcom.wasm ./cmd/termcom_wasm/

rm -rf "$WASMDATA"

echo "Compressing WASM binary..."
gzip -k web_wasm/termcom.wasm

WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ -f "$WASM_EXEC" ]; then
	cp "$WASM_EXEC" web_wasm/
else
	WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
	if [ -f "$WASM_EXEC" ]; then
		cp "$WASM_EXEC" web_wasm/
	else
		echo "WARNING: wasm_exec.js not found at expected paths"
	fi
fi

echo "WASM build complete: web_wasm/"
echo "  termcom.wasm       ($(du -h web_wasm/termcom.wasm | cut -f1))"
echo "  termcom.wasm.gz    ($(du -h web_wasm/termcom.wasm.gz | cut -f1))"
echo "  wasm_exec.js"
echo "  index.html"
echo ""
echo "To test locally: cd web_wasm && python3 -m http.server 8080"
