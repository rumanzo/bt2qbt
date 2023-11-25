gotag=1.20.3-bullseye

commit=$(shell git rev-parse HEAD)

dockercmd=docker run --rm -v $(CURDIR):/usr/src/bt2qbt -w /usr/src/bt2qbt
buildtags = -tags forceposix
buildenvs = -e CGO_ENABLED=0
version = 1.999
ldflags = -ldflags="-X 'main.version=$(version)' -X 'main.commit=$(commit)' -X 'main.buildImage=golang:$(gotag)'"

all: | tests build

tests:
	$(dockercmd) golang:$(gotag) go test $(buildtags) ./...

build: windows linux darwin

windows:
	$(dockercmd) $(buildenvs) -e GOOS=windows -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64.exe
	$(dockercmd) $(buildenvs) -e GOOS=windows -e GOARCH=386 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_i386.exe

linux:
	$(dockercmd) $(buildenvs) -e GOOS=linux -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64_linux
	$(dockercmd) $(buildenvs) -e GOOS=linux -e GOARCH=386 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_i386_linux

darwin:
	$(dockercmd) $(buildenvs) -e GOOS=darwin -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64_macos
	$(dockercmd) $(buildenvs) -e GOOS=darwin -e GOARCH=arm64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_arm64_macos
