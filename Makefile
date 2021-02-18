SHELL := bash
.ONESHELL:

VER=$(shell git describe --tags --always --dirty)
GO=$(shell which go)
GOGET=$(GO) get
GOMOD=$(GO) mod
GOFMT=$(GO) fmt
GOBUILD=$(GO) build -mod=readonly -ldflags "-X main.version=$(VER)"

dir:
	@if [ ! -d bin ]; then mkdir -p bin; fi

mod:
	@$(GOMOD) download

format:
	@$(GOFMT) ./...

build/linux/armv7: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=arm
	export GOARM=7
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-armv7

build/linux/arm64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=arm64
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-arm64

build/linux/i386: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=386
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-i386

build/linux/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=amd64
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-amd64

build/linux: build/linux/armv7 build/linux/arm64 build/linux/i386 build/linux/amd64

build/darwin/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=darwin
	export GOARCH=amd64
	$(GOBUILD) -o bin/go-lg-darwin-$(VER:v%=%)-amd64

build/darwin: build/darwin/amd64

build/windows/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=windows
	export GOARCH=amd64
	$(GOBUILD) -o bin/go-lg-windows-$(VER:v%=%)-amd64

build/windows: build/windows/amd64

build: build/linux build/darwin build/windows

clean:
	@rm -rf bin
	@rm -rf internal/assets

all: format build
