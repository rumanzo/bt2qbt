#!/bin/bash
GOOS=windows GOACH=amd64 go build -o bt2qbt_v${1}_amd64.exe
GOOS=linux GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_linux
GOOS=darwin GOARCH=amd64 go build -o bt2qbt_v${1}_amd64_macos
GOOS=linux GOARCH=386 go build -o bt2qbt_v${1}_i386_linux
GOOS=linux GOARCH=386 go build -o bt2qbt_v${1}_i386.exe
