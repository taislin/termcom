.PHONY: build test lint coverage clean run

export PATH := /home/taislin/go/bin:$(PATH)
export GOPATH := /home/taislin/gopath

LDFLAGS := -ldflags="-X github.com/taislin/termcom/internal/engine.GameVersion=$(shell cat VERSION)"

run:
	go run $(LDFLAGS) ./cmd/termcom

build:
	go build $(LDFLAGS) -o termcom ./cmd/termcom

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	@echo ""
	@echo "HTML report: go tool cover -html=coverage.out"

lint:
	go vet ./...
	which staticcheck || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

clean:
	rm -f termcom coverage.out
