SHELL := bash
.ONESHELL:

VER=$(shell git describe --tags --always --dirty)
GO=$(shell which go)
GOOPTS=-trimpath -mod=readonly -ldflags "-X main.version=$(VER:v%=%) -s -w -buildid="
GOINSTALL=$(GO) install
GOMOD=$(GO) mod
GOFMT=$(GO) fmt
GOBUILD=$(GO) build $(GOOPTS)

dir:
	@if [ ! -d bin ]; then mkdir -p bin; fi

mod:
	@$(GOMOD) download
	@$(GOINSTALL) github.com/google/go-licenses@latest

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

build/linux/386: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=386
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-386

build/linux/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=amd64
	$(GOBUILD) -o bin/go-lg-linux-$(VER:v%=%)-amd64

build/linux: build/linux/armv7 build/linux/arm64 build/linux/386 build/linux/amd64

build: build/linux

license: dir
	cp NOTICE bin/NOTICE
	cp LICENSE bin/LICENSE
	go-licenses save . --ignore github.com/CHTJonas/go-lg --save_path bin/licenses
	(cd bin/licenses && zip -r ../third-party-licenses.zip . -i **/LICENSE **/LICENSE.* **/NOTICE **/NOTICE.* **/COPYING **/COPYING.*)
	rm -rf bin/licenses

clean:
	@rm -rf bin

all: format build license
