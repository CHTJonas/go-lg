SHELL := bash
.ONESHELL:

GO=$(shell which go)
GOGET=$(GO) get
GOFMT=$(GO) fmt
GOBUILD=$(GO) build -ldflags "-X main.ver=`git describe --tags`"

BINDATA=$(shell which go-bindata)

dir:
	@if [ ! -d bin ] ; then mkdir -p bin ; fi

get:
	$(GOGET) github.com/shuLhan/go-bindata/...
	$(GOGET) github.com/dgraph-io/badger
	$(GOGET) github.com/cbroglie/mustache
	$(GOGET) github.com/gorilla/mux

format:
	$(GOFMT) ./...

build/assets:
	go-bindata -o internal/assets/assets.go -pkg assets assets/

build/linux/armv7: build/assets
	export GOOS=linux
	export GOARCH=arm
	export GOARM=7
	$(GOBUILD) -o bin/linux-armv7/go-lg cmd/go-lg/main.go

build/linux/arm64: build/assets
	export GOOS=linux
	export GOARCH=arm64
	$(GOBUILD) -o bin/linux-arm64/go-lg cmd/go-lg/main.go

build/linux/i386: build/assets
	export GOOS=linux
	export GOARCH=386
	$(GOBUILD) -o bin/linux-i386/go-lg cmd/go-lg/main.go

build/linux/amd64: build/assets
	export GOOS=linux
	export GOARCH=amd64
	$(GOBUILD) -o bin/linux-amd64/go-lg cmd/go-lg/main.go

build/linux: build/linux/armv7 build/linux/arm64 build/linux/i386 build/linux/amd64

build/darwin/amd64: build/assets
	export GOOS=darwin
	export GOARCH=amd64
	$(GOBUILD) -o bin/darwin-amd64/go-lg cmd/go-lg/main.go

build/darwin: build/darwin/amd64

build/windows/amd64:
	export GOOS=windows
	export GOARCH=amd64
	$(GOBUILD) -o bin/windows-amd64/go-lg cmd/go-lg/main.go

build/windows: build/windows/amd64

build: build/linux build/darwin build/windows

clean:
	@rm -rf bin
	@rm -rf internal/assets

all: dir format build
