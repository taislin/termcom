go build -ldflags="-s -w" -trimpath ./cmd/termcom
upx --best --lzma termcom