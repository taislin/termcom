//go:build js && wasm

package main

import (
	"embed"
	"io/fs"

	"github.com/taislin/termcom/internal/datafs"
)

//go:embed wasmdata
var wasmData embed.FS

func init() {
	datafs.Set(embeddedFS())
}

func embeddedFS() fs.FS {
	sub, err := fs.Sub(wasmData, "wasmdata")
	if err != nil {
		return wasmData
	}
	return sub
}
