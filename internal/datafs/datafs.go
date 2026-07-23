package datafs

import (
	"io/fs"
	"os"
	"path/filepath"
)

var FS fs.FS = os.DirFS(".")

func Set(fsys fs.FS) { FS = fsys }

func ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(FS, path)
}

func ReadDir(path string) ([]fs.DirEntry, error) {
	return fs.ReadDir(FS, path)
}

func Glob(pattern string) ([]string, error) {
	return fs.Glob(FS, filepath.ToSlash(pattern))
}
