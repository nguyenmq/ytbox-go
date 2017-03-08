/*
 * Entry point into the yt_box backend service
 */

package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/nguyenmq/ytbox-go/backend"
	"github.com/nguyenmq/ytbox-go/common"
)

/*
 * Command line arguments
 */
var (
	app      = kingpin.New(backend.LogPrefix, "yt_box backend server")
	host     = app.Flag("host", "Address to listen on").Default("127.0.0.1").Short('h').String()
	port     = app.Flag("port", "Port to listen on").Default("8000").Short('p').String()
	loadFile = app.Flag("load", "Load a serialized protobuf playlist from a file").Short('l').ExistingFile()
)

func main() {
	app.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logFile := common.InitLogger(backend.LogPrefix, true)
	defer logFile.Close()

	ytbServer := backend.NewServer(*host+":"+*port, *loadFile)
	ytbServer.Serve()
}
