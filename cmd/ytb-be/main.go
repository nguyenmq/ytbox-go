// Entry point into the yt_box backend service

package main

import (
	"github.com/nguyenmq/ytbox-go/backend"
	"github.com/nguyenmq/ytbox-go/common"
)

func main() {
	logFile := ytb_common.InitLogger(ytbbe.LogPrefix, true)
	defer logFile.Close()

	ytbServer := ytbbe.NewServer(":8000")
	ytbServer.Serve()
}
