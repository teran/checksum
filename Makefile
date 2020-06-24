export PACKAGES := $(shell env GOPATH=$(GOPATH) go list ./...)
export VERSION := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') || git rev-parse --verify --short HEAD || echo ${VERSION})
export COMMIT := $(shell git rev-parse --verify --short HEAD)
export DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

all: clean build

clean:
	rm -vf bin/*

build: build-macos build-linux build-windows

build-macos: build-macos-amd64 build-macos-i386

build-linux: build-linux-amd64 build-linux-i386

build-windows: build-windows-amd64 build-windows-i386

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-darwin-amd64 .

build-macos-i386:
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-darwin-i386 .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-linux-amd64 .

build-linux-i386:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-linux-i386 .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-windows-amd64.exe .

build-windows-i386:
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/checksum-windows-i386.exe .

sign:
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-darwin-amd64.sig 				bin/checksum-darwin-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-darwin-i386.sig 				bin/checksum-darwin-i386
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-linux-amd64.sig 				bin/checksum-linux-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-linux-i386.sig 					bin/checksum-linux-i386
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-windows-amd64.exe.sig 	bin/checksum-windows-amd64.exe
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/checksum-windows-i386.exe.sig 		bin/checksum-windows-i386.exe

test:
	go test ./...

verify:
	gpg --verify bin/checksum-darwin-amd64.sig 				bin/checksum-darwin-amd64
	gpg --verify bin/checksum-darwin-i386.sig 				bin/checksum-darwin-i386
	gpg --verify bin/checksum-linux-amd64.sig 				bin/checksum-linux-amd64
	gpg --verify bin/checksum-linux-i386.sig 					bin/checksum-linux-i386
	gpg --verify bin/checksum-windows-amd64.exe.sig 	bin/checksum-windows-amd64.exe
	gpg --verify bin/checksum-windows-i386.exe.sig 		bin/checksum-windows-i386.exe
