## Overview
Acts like a virtual jukebox where users can submit YouTube links to a web
frontend and a player application will then play the songs in the playlist.

## Setup
Install [Go](https://golang.org/doc/install). Set up your [workspace and GOPATH](https://golang.org/doc/code.html).

Use `go get -d ./...` from the root of the repo to grab all dependencies

Install [protobuf](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)
and the Go plugin. Build the `.proto` files:
```
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I proto/backend -I "$GOPATH/src" --go_out=plugins=grpc:proto/backend proto/backend/*.proto
protoc -I proto/common -I "$GOPATH/src" --go_out=plugins=grpc:proto/common proto/common/common.proto
```

Install `youtube-dl` and `mpv`.

## Build
The `cmd` sub-directory contains several binaries that can be built using `go
build` or `go install`.

### Examples
Building a single command with debug information and place it locally within
a `bin` directory:
```
go build -o bin/cli-be ./cmd/ytb-be-cli
go build -o bin/backend ./cmd/ytb-be
go build -o bin/frontend ./cmd/ytb-fe
go build -o bin/player ./cmd/ytb-player/
```
