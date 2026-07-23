//go:build js && wasm

package main

import (
	"embed"
	"io/fs"
)

//go:embed wasmdata
var wasmData embed.FS

func embeddedFS() fs.FS {
	sub, err := fs.Sub(wasmData, "wasmdata")
	if err != nil {
		return wasmData
	}
	return sub
}
