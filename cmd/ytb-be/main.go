/*
 * Entry point into the yt_box backend service
 */

package main

import (
	"log"
	"os"
	"os/signal"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/nguyenmq/ytbox-go/backend"
	"github.com/nguyenmq/ytbox-go/common"
)

/*
 * Command line arguments
 */
var (
	app      = kingpin.New(backend.LogPrefix, "yt_box backend server")
	all      = app.Flag("all", "Listen on all interfaces. Only listens on localhost by default.").Short('a').Bool()
	port     = app.Flag("port", "Port to listen on").Default("8000").Short('p').String()
	loadFile = app.Flag("load", "Load a serialized protobuf playlist from a file").Short('l').ExistingFile()
	dbFile   = app.Flag("database", "Path to database").Default("./ytbox.db").Short('d').String()
)

func main() {
	app.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logFile := common.InitLogger(backend.LogPrefix, true)
	defer logFile.Close()

	addr := "127.0.0.1"
	if *all {
		addr = "0.0.0.0"
	}

	ytbServer := backend.NewServer(addr+":"+*port, *loadFile, *dbFile)

	go func() {
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt)

		select {
		case <-stop:
			ytbServer.Stop()
		}
	}()

	log.Println("Server started")
	ytbServer.Serve()
	log.Println("Server stopped")
}
