SHELL := bash
.ONESHELL:

GO=$(shell which go)
GOGET=$(GO) get
GOFMT=$(GO) fmt
GOBUILD=$(GO) build

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

build/armv7: build/assets
	export GOOS=linux
	export GOARCH=arm
	export GOARM=7
	$(GOBUILD) -o bin/go-lg-armv7 cmd/go-lg/main.go

build/arm64: build/assets
	export GOOS=linux
	export GOARCH=arm64
	$(GOBUILD) -o bin/go-lg-arm64 cmd/go-lg/main.go

build/386: build/assets
	export GOOS=linux
	export GOARCH=386
	$(GOBUILD) -o bin/go-lg-386 cmd/go-lg/main.go

build/amd64: build/assets
	export GOOS=linux
	export GOARCH=amd64
	$(GOBUILD) -o bin/go-lg-amd64 cmd/go-lg/main.go

build: build/armv7 build/arm64 build/386 build/amd64

clean:
	@rm -rf bin
	@rm -rf internal/assets

all: dir format build
