.PHONY: build test lint coverage clean run

export PATH := /home/civ13/go/bin:$(PATH)
export GOPATH := /home/civ13/gopath

run:
	go run ./cmd/termcom

build:
	go build -o termcom ./cmd/termcom

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
