package main

import (
	"fmt"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/nguyenmq/ytbox-go/common"
	pb "github.com/nguyenmq/ytbox-go/proto"
)

const (
	prefix string = "ytb-be-cli"
)

/*
 * Command line arguments
 */
var (
	app        = kingpin.New(prefix, "Command line client to ytb-be")
	remoteHost = app.Flag("host", "Address of remote ytb-be service").Default("127.0.0.1").Short('h').String()
	remotePort = app.Flag("port", "Port of remote ytb-be service").Default("8000").Short('p').String()

	// "submit" subcommand
	submit     = app.Command("submit", "Submit a link to the queue.")
	submitLink = submit.Flag("link", "Link to song.").Short('l').Required().String()
	submitUser = submit.Flag("user", "User id to submit link under.").Short('u').Required().Uint32()

	// "playlist" subcommand
	playlist = app.Command("playlist", "Get current songs in the playlist")
)

/*
 * Connect to the remote server. Remember to close the returned connect when
 * done.
 */
func connectToRemote() (*grpc.ClientConn, pb.YtbBackendClient) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*remoteHost+":"+*remotePort, opts...)
	if err != nil {
		fmt.Printf("failed to dial server: %v\n", err)
		os.Exit(1)
	}

	client := pb.NewYtbBackendClient(conn)
	return conn, client
}

/*
 * Handler to submit a song to the remote server
 */
func submitCommand() {
	conn, client := connectToRemote()
	defer conn.Close()

	response, err := client.SubmitSong(context.Background(), &pb.SubmitMessage{*submitLink, *submitUser})
	if err != nil {
		fmt.Printf("failed to call SubmitSong: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: {flag: %t, message: %s}\n", response.Success, response.Message)
}

/*
 * Handler to list the songs in the playlist
 */
func playlistCommand() {
	conn, client := connectToRemote()
	defer conn.Close()

	playlist, err := client.GetPlaylist(context.Background(), &pb.Empty{})
	if err != nil {
		fmt.Printf("failed to call GetPlaylist: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Songs in the playlist:\n")
	for i := 0; i < len(playlist.Songs); i++ {
		fmt.Printf("%3d. %s\n", i+1, playlist.Songs[i].Title)
	}
}

func main() {
	logFile := common.InitLogger(prefix, true)
	defer logFile.Close()

	kingpin.Version("0.1")

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case submit.FullCommand():
		submitCommand()

	case playlist.FullCommand():
		playlistCommand()
	}
}
