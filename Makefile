#GOOS=windows GOACH=amd64 go build -o bt2qbt_v${1}_amd64.exe -tags forceposix
#GOOS=windows GOARCH=386 go build -o bt2qbt_v${1}_i386.exe -tags forceposix
#GOOS=linux GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_linux -tags forceposix
#GOOS=linux GOARCH=386 go build -o bt2qbt_v${1}_i386_linux -tags forceposix
#GOOS=darwin GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_macos -tags forceposix
#GOOS=darwin GOARCH=386 go build -o bt2qbt_v${1}_i386_macos -tags forceposix

GOVER=1.18.0
GOTAG=1.18.0-bullseye
OLDGOVER=1.15.15
OLDGOTAG=1.15.15-buster

CGO_ENABLED?=0

COMMIT=$(shell git rev-parse HEAD)

DOCKERCMD=docker run --rm -v $(CURDIR):/usr/src/bt2qbt -w /usr/src/bt2qbt
BUILDTAGS=-tags forceposix


all: tests build

tidy:
	$(DOCKERCMD) golang:$(GOTAG) go mod tidy

tests:
	$(DOCKERCMD) golang:$(GOTAG) go test $(BUILDTAGS) ./pkg/fileHelpers

build: tests
	$(DOCKERCMD) golang:$(GOTAG) go build $(BUILDTAGS) ./cmd/bt2qbt/
