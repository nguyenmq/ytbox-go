package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"

	mpv "github.com/DexterLB/mpvipc"
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

const (
	mpvSocket = "./.mpvsocket"
)

/*
 * Remote contoller to interface with mpv
 */
type Remote struct {
	conn *mpv.Connection
}

/*
 * Initialize the remote
 */
func (r *Remote) Init(conn *mpv.Connection) {
	r.conn = conn
}

/*
 * Load a song into mpv
 */
func (r *Remote) LoadSong(name string) {
	_, err := r.conn.Call("loadfile", name, "append-play")
	if err != nil {
		fmt.Printf("Failed to load song: %v\n", err)
	}
}

/*
 * Toggle the pause state
 */
func (r *Remote) TogglePause() {
	_, err := r.conn.Call("cycle", "pause", "up")
	if err != nil {
		fmt.Printf("Failed to toggle pause: %v\n", err)
	}
}

/*
 * Tell mpv to quit
 */
func (r *Remote) Quit() {
	if !r.conn.IsClosed() {
		_, err := r.conn.Call("quit")
		if err != nil {
			fmt.Printf("Failed to call quit: %v\n", err)
		}
	}
}

/*
 * Connect to the remote server and create an RPC client
 */
func connectToRemote() (*grpc.ClientConn, bepb.YtbBePlayer_SongPlayerClient) {
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.FailOnNonTempDialError(true))

	conn, err := grpc.Dial(*remoteHost+":"+*remotePort, opts...)
	if err != nil {
		fmt.Printf("failed to dial server: %v\n", err)
		os.Exit(1)
	}

	client := bepb.NewYtbBePlayerClient(conn)
	stream, err := client.SongPlayer(context.Background())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}

	return conn, stream
}

/*
 * Receieve a new status and song from the server.
 */
func receiveStatus(stream bepb.YtbBePlayer_SongPlayerClient, newStatus chan *bepb.PlayerControl) {
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

		newStatus <- status
	}
}

/*
 * Handle a new status message from the server.
 */
func handleNewStatus(status *bepb.PlayerControl, remote *Remote) {
	fmt.Printf("Received: %v\n", status)

	if status.Command == bepb.CommandType_Play {
		playSong(status.Song, remote)
	}
}

/*
 * Handle messages from other goroutines.
 */
func interactionLoop(stream bepb.YtbBePlayer_SongPlayerClient, conn *mpv.Connection) {
	newStatus := make(chan *bepb.PlayerControl)
	halt := make(chan os.Signal)
	mpvExit := make(chan struct{})
	signal.Notify(halt, os.Interrupt)
	remote := new(Remote)
	remote.Init(conn)
	events, stop := conn.NewEventListener()
	streamOk := true
	running := true

	// send the initial command to the server to signal the player is ready
	stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})

	// start receiving messages
	go receiveStatus(stream, newStatus)

	// if mpv exits before player, then signal an exit to player
	go func() {
		conn.WaitUntilClosed()
		close(mpvExit)
	}()

	for running {
		select {
		case status, streamOk := <-newStatus:
			if !streamOk {
				running = false
				break
			}
			handleNewStatus(status, remote)

		case <-mpvExit:
			running = false
			break

		case <-halt:
			running = false
			break

		case event := <-events:
			if event.Name == "idle" {
				stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})
			}
		}
	}

	// tell the server we're exiting
	if streamOk {
		stream.CloseSend()
	}

	// clean up mpv connection
	time.Sleep(100 * time.Millisecond)
	if !conn.IsClosed() {
		stop <- struct{}{}
		remote.Quit()
	}
	conn.Close()

	close(halt)
}

/*
 * Build the song link and play the given song
 */
func playSong(song *cmpb.Song, remote *Remote) {
	var link string

	switch song.Service {
	case cmpb.ServiceType_ServiceLocal:
		link = song.ServiceId

	case cmpb.ServiceType_ServiceYoutube:
		link = fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ServiceId)

	default:
		fmt.Printf("Unsupported link: %s\n", song.ServiceId)
	}

	remote.LoadSong(link)
}

/*
 * Start mpv in idle mode
 */
func startMpv() *exec.Cmd {
	socketFlag := "--input-ipc-server=" + mpvSocket
	cmd := exec.Command("mpv", "--idle", socketFlag, "--fullscreen", "--force-window")
	cmd.Start()
	time.Sleep(500 * time.Millisecond)
	return cmd
}

func main() {
	kingpin.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	conn, stream := connectToRemote()
	defer conn.Close()

	mpvCmd := startMpv()

	mpvConn := mpv.NewConnection(mpvSocket)
	if err := mpvConn.Open(); err != nil {
		panic(err)
	}

	interactionLoop(stream, mpvConn)

	mpvCmd.Wait()
	os.Remove(mpvSocket)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("end")
}
