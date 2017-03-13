## Overview
Acts like a virtual jukebox where users can submit YouTube links to a web
frontend and a player application will then play the songs in the playlist.

## Setup
Install [Go](https://golang.org/doc/install). Set up your [workspace and GOPATH](https://golang.org/doc/code.html).

Use `go get` to obtain the following libraries:
- github.com/dhowden/tag
- github.com/golang/protobuf/proto
- github.com/mattn/go-sqlite3
- golang.org/x/net/context
- google.golang.org/grpc
- gopkg.in/alecthomas/kingpin.v2

Install [protobuf](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)
and the Go plugin. Build the `.proto` files:
```
protoc -I proto/backend --go_out=plugins=grpc:proto/backend proto/backend/backend.proto
```

Install `youtube-dl` and `mpv`.

## Build
The `cmd` sub-directory contains several binaries that can be built using `go
build` or `go install`.

### Examples
Building a single command with debug information and place it locally within
a `bin` directory:
```
go build -o bin/cli-be -gcflags "-N -l" ./cmd/ytb-be-cli
```
