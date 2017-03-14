package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

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

func sendStatus(stream bepb.YtbBePlayer_SongPlayerClient, playing chan int) bool {
	<-playing
	stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})
	return false
}

func receiveStatus(stream bepb.YtbBePlayer_SongPlayerClient, playing chan int) bool {
	retry := true

	for retry {
		status, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("Disconnected")
			return true
		}

		if err != nil {
			fmt.Printf("failed to receive controller messager: %v\n", err)
			return true
		}

		fmt.Printf("Received: %v\n", status)

		if status.Command == bepb.CommandType_Play {
			go playSong(status.Song, playing)
			retry = false
		} else {
			retry = true
		}

		return false
	}

	return false
}

func playSong(song *cmpb.Song, playing chan int) {
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

	playing <- 1
}

func main() {
	kingpin.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	send := make(chan int)
	receive := make(chan int)
	playing := make(chan int)
	stop := false

	conn, client := connectToRemote()
	defer conn.Close()

	stream, err := client.SongPlayer(context.Background())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}

	// receive a message from the server
	go func() {
		playing <- 1
		for !stop {
			<-receive
			stop = receiveStatus(stream, playing)
			send <- 1
		}
	}()

	sendStatus(stream, playing)
	receive <- 1

	// send a message to the server
	for !stop {
		<-send
		stop = sendStatus(stream, playing)
		receive <- 1
	}

	fmt.Println("end")
}
