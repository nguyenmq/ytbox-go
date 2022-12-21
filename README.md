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
make proto
```

Install `yt-dlp` and `mpv`.

## Build
The `cmd` sub-directory contains several binaries that can be built using `go
build` or `go install`.

### Examples

See `Makefile` for build targets.

Might still need to build the backend with sqlite3 explicitly defined.

```
go build --tags "libsqlite3 linux" -v -o bin/backend ./cmd/ytb-be
```
