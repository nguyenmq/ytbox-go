package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * Command line arguments
 */
var (
	app        = kingpin.New("ytb-player", "Command line client to play videos in the ytb-be queue")
	remoteHost = app.Flag("host", "Address of remote ytb-be service").Default("127.0.0.1").Short('h').String()
	remotePort = app.Flag("port", "Port of remote ytb-be service").Default("8000").Short('p').String()
	continuous = app.Flag("cont", "Continuous play songs from the queue").Short('c').Bool()
)

/*
 * Connect to the remote server. Remember to close the returned connect when
 * done.
 */
func connectToRemote() (*grpc.ClientConn, bepb.YtbBePlayerClient) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*remoteHost+":"+*remotePort, opts...)
	if err != nil {
		fmt.Printf("failed to dial server: %v\n", err)
		os.Exit(1)
	}

	client := bepb.NewYtbBePlayerClient(conn)
	return conn, client
}

/*
 * Receieve a new status and song from the server.
 */
func receiveSong(stream bepb.YtbBePlayer_SongPlayerClient, newStatus chan bepb.PlayerControl) {
	for {
		status, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("Disconnected")
			close(newStatus)
			break
		}

		if err != nil {
			fmt.Printf("failed to receive controller messager: %v\n", err)
			close(newStatus)
			break
		}

		newStatus <- *status
	}
}

/*
 * Handle a new status message from the server.
 */
func handleNewStatus(status *bepb.PlayerControl, stream bepb.YtbBePlayer_SongPlayerClient) {
	fmt.Printf("Received: %v\n", status)

	if status.Command == bepb.CommandType_Play {
		go playSong(status.Song, stream)
	}
}

/*
 * Handle messages from other go routines.
 */
func interactionLoop(stream bepb.YtbBePlayer_SongPlayerClient) {
	isRunning := true
	newStatus := make(chan bepb.PlayerControl)
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	// send the initial command to the server to signal the player is ready
	stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})

	go receiveSong(stream, newStatus)

	for isRunning {
		select {
		case status, ok := <-newStatus:
			if !ok {
				return
			}
			handleNewStatus(&status, stream)

		case <-stop:
			stream.CloseSend()
			return
		}

	}
}

/*
 * Play the given song and return a ready status to the remote server when
 * done.
 */
func playSong(song *cmpb.Song, stream bepb.YtbBePlayer_SongPlayerClient) {
	var link string

	switch song.Service {
	case cmpb.ServiceType_ServiceLocal:
		link = song.ServiceId

	case cmpb.ServiceType_ServiceYoutube:
		link = fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ServiceId)

	default:
		fmt.Printf("Unsupported link: %s\n", song.ServiceId)
	}

	if link != "" {
		err := exec.Command("mpv", "--fs", link).Run()
		if err != nil {
			fmt.Printf("Failed to play link: %s\n", link)
		}
	}

	// tell the remote server the player is ready for another song
	stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})
}

func main() {
	kingpin.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	conn, client := connectToRemote()
	defer conn.Close()

	stream, err := client.SongPlayer(context.Background())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}

	interactionLoop(stream)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("end")
}
