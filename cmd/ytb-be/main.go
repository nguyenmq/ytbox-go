// Entry point into the yt_box backend service

package main

import (
	"github.com/nguyenmq/ytbox-go/backend"
	"github.com/nguyenmq/ytbox-go/common"
)

func main() {
	logFile := common.InitLogger(backend.LogPrefix, true)
	defer logFile.Close()

	ytbServer := backend.NewServer(":8000")
	ytbServer.Serve()
}
