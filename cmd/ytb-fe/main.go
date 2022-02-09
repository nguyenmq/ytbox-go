/*
 * Entry point into the yt_box backend service
 */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/nguyenmq/ytbox-go/internal/common"
	"github.com/nguyenmq/ytbox-go/internal/frontend"
)

/*
 * Command line arguments
 */
var (
	app       = kingpin.New(frontend.LogPrefix, "yt_box frontend server")
	all       = app.Flag("all", "Listen on all interfaces. Only listens on localhost by default.").Short('a').Bool()
	port      = app.Flag("port", "Port to listen on").Default("9008").Short('p').String()
	hashFile  = app.Flag("hash", "File containing hash key").Default("hash.key").String()
	blockFile = app.Flag("block", "File containing block key").Default("block.key").String()
	debug     = app.Flag("debug", "Enable debug mode.").Short('d').Bool()
)

func main() {
	app.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logFile := common.InitLogger(frontend.LogPrefix, true)
	defer logFile.Close()

	addr := "127.0.0.1"
	if *all {
		addr = "0.0.0.0"
	}

	hashKey, err := ioutil.ReadFile(*hashFile)
	if err != nil {
		log.Printf("Could not read hash key file at %s with error: %s\n", *hashFile, err.Error())
		os.Exit(1)
	}

	blockKey, err := ioutil.ReadFile(*blockFile)
	if err != nil {
		log.Printf("Could not read block key file at %s with error: %s\n", *blockFile, err.Error())
		os.Exit(1)
	}

	server := frontend.NewServer(addr+":"+*port, []byte(hashKey), []byte(blockKey), *debug)

	go func() {
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt)

		select {
		case <-stop:
			server.Stop()
		}
	}()

	log.Println("Server started")
	server.Start()
	log.Println("Server stopped")
}
