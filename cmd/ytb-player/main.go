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

	bepb "github.com/nguyenmq/ytbox-go/internal/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
)

/*
 * Command line arguments
 */
var (
	app        = kingpin.New("ytb-player", "Command line client to play videos in the ytb-be queue")
	remoteHost = app.Flag("host", "Address of remote ytb-be service").Default("127.0.0.1").Short('h').String()
	remotePort = app.Flag("port", "Port of remote ytb-be service").Default("9009").Short('p').String()
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
func (r *Remote) LoadSong(name string, play bool) {
	var err error
	if play {
		_, err = r.conn.Call("loadfile", name, "append-play")
	} else {
		_, err = r.conn.Call("loadfile", name, "append")
	}

	if err != nil {
		fmt.Printf("Failed to load song: %v\n", err)
	}
}

/*
 * Show the given text on mpv's OSD. Duration is given in milliseconds.
 */
func (r *Remote) ShowText(text string, duration string) {
	_, err := r.conn.Call("show-text", text, duration, 1)
	if err != nil {
		fmt.Printf("Failed to show text: %v\n", err)
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
 * Force pause to a certain state
 */
func (r *Remote) ForcePause(state bool) {
	_, err := r.conn.Call("set_property", "pause", state)
	if err != nil {
		fmt.Printf("Failed to pause: %v\n", err)
	}
}

/*
 * Get the number of tracks in mpv's playlist
 */
func (r *Remote) GetPlaylistCount() float64 {
	count, err := r.conn.Get("playlist-count")
	if err != nil {
		fmt.Printf("Failed to pause: %v\n", err)
		return 0
	}

	return count.(float64)
}

/*
 * Go to the next song
 */
func (r *Remote) Next(name string) {
	if name != "" {
		r.LoadSong(name, true)
	}

	// a new player should have one track in the playlist. Players who are
	// already playing should have two
	if r.GetPlaylistCount() > 1.0 {
		_, err := r.conn.Call("playlist-next", "force")
		if err != nil {
			fmt.Printf("Failed to go to next song: %v\n", err)
		} else {
			r.ForcePause(false)
		}
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

	switch status.GetCommand() {
	case bepb.CommandType_Play:
		link, ok := buildSongLink(status.GetSong())
		if ok {
			remote.LoadSong(link, true)
			remote.ShowText(status.GetSong().GetTitle(), "8000")
		}

	case bepb.CommandType_Next:
		// link can be an empty string. We still want to stop the player even
		// if there are no more songs in the playlist
		link, _ := buildSongLink(status.GetSong())
		remote.Next(link)

	case bepb.CommandType_Pause:
		remote.TogglePause()
		remote.ShowText("Player is paused", "600000")
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

	remote.ShowText("Waiting for users to add songs", "600000")

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
				remote.ShowText("Waiting for users to add songs", "600000")
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
 * Build the song link
 */
func buildSongLink(song *cmpb.Song) (string, bool) {
	link := ""
	ok := true

	switch song.GetService() {
	case cmpb.ServiceType_Local:
		link = song.GetServiceId()

	case cmpb.ServiceType_Youtube:
		link = fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.GetServiceId())

	case cmpb.ServiceType_None:
		ok = false

	default:
		fmt.Printf("Unsupported link: %s\n", song.GetServiceId())
		ok = false
	}

	return link, ok
}

/*
 * Start mpv in idle mode
 */
func startMpv() *exec.Cmd {
	socketFlag := "--input-ipc-server=" + mpvSocket
	cmd := exec.Command("mpv",
		"--idle",
		socketFlag,
		"--fullscreen",
		"--force-window",
		"--no-osc",
		"--osd-align-y=bottom",
		"--osd-blur=1.0",
		"--osd-font='Ubuntu'",
		"--osd-font-size=44")
	//"--audio-delay=-1.3",
	//"--audio-display=no",
	//"--audio-channels=stereo",
	//"--audio-samplerate=48000",
	//"--audio-format=s16",
	//"--ao=pcm",
	//"--ao-pcm-file=/tmp/snapfifo")

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
