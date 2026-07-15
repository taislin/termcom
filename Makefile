.PHONY: build test lint coverage clean run android-aar android-apk android-install

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
	which staticcheck || go install honnef.co/go/tools/cmd/staticcheck@v0.4.0
	staticcheck ./...

clean:
	rm -f termcom coverage.out
	rm -rf android/app/build android/app/libs

android-aar:
	gomobile bind -target=android -androidapi 21 -o android/app/libs/termcom.aar -ldflags="-X github.com/taislin/termcom/internal/engine.GameVersion=$(shell cat VERSION)" ./android/

android-apk: android-aar
	cd android && gradle assembleDebug

android-install: android-apk
	cd android && gradle installDebug
