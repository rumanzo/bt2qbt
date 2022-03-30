#GOOS=windows GOACH=amd64 go build -o bt2qbt_v${1}_amd64.exe -tags forceposix
#GOOS=windows GOARCH=386 go build -o bt2qbt_v${1}_i386.exe -tags forceposix
#GOOS=linux GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_linux -tags forceposix
#GOOS=linux GOARCH=386 go build -o bt2qbt_v${1}_i386_linux -tags forceposix
#GOOS=darwin GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_macos -tags forceposix
#GOOS=darwin GOARCH=386 go build -o bt2qbt_v${1}_i386_macos -tags forceposix

gotag=1.18.0-bullseye

commit=$(shell git rev-parse HEAD)

dockercmd=docker run --rm -v $(CURDIR):/usr/src/bt2qbt -w /usr/src/bt2qbt
buildtags = -tags forceposix
buildenvs = -e CGO_ENABLED=0
version = 1999
ldflags = -ldflags="-X 'main.version=$(version)' -X 'main.commit=$(commit)' -X 'main.buildImage=golang:$(gotag)'"

all: tests build

tests:
	$(dockercmd) golang:$(gotag) go test $(buildtags) ./...

build: | tests windows linux darwin


windows:
	$(dockercmd) $(buildenvs) -e GOOS=windows -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64.exe
	$(dockercmd) $(buildenvs) -e GOOS=windows -e GOARCH=386 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_i386.exe

linux:
	$(dockercmd) $(buildenvs) -e GOOS=linux -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64_linux
	$(dockercmd) $(buildenvs) -e GOOS=linux -e GOARCH=386 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_i386_linux

darwin:
	$(dockercmd) $(buildenvs) -e GOOS=darwin -e GOARCH=amd64 golang:$(gotag) go build -v $(buildtags) $(ldflags) -o bt2qbt_v$(version)_amd64_macos
