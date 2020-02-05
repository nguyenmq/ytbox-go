package main

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	prefix string = "ytb-be-cli"
)

/*
 * Command line arguments
 */
var (
	app        = kingpin.New(prefix, "Command line client to ytb-be.")
	remoteHost = app.Flag("host", "Address of remote ytb-be service.").Default("127.0.0.1").Short('h').String()
	remotePort = app.Flag("port", "Port of remote ytb-be service.").Default("9009").Short('p').String()

	// "playlist" subcommand
	playlist = app.Command("playlist", "Get current songs in the playlist.").Alias("ls")

	// "login" subcommand
	login       = app.Command("login", "Login as the given username.")
	loginName   = login.Arg("username", "Alias to login as.").Required().String()
	loginRoomId = login.Arg("roomId", "Id of the room to log user into.").Required().Uint32()
	loginId     = login.Arg("userId", "Id of the alias to login as.").Uint32()

	// "next" subcommand
	next = app.Command("next", "Skip to the next song.")

	// "now" subcommand
	now = app.Command("now", "Get the current song that is playing.").Default()

	// "pause" subcommand
	pause = app.Command("pause", "Toggle pause state of the player.")

	// "pop" subcommand
	pop = app.Command("pop", "Pop a song off the top of the queue.")

	// "remove" subcommand
	remove     = app.Command("remove", "Remove a song from the playlist.").Alias("rm")
	removeSong = remove.Arg("songId", "Id of the song to remove.").Required().Uint32()
	removeUser = remove.Arg("userId", "Id of the user who subitted the song.").Required().Uint32()

	// "save" subcommand
	save     = app.Command("save", "Save the current playlist to a file.")
	saveFile = save.Arg("file", "File name to write playlist to").Required().String()

	// "send" subcommand
	send     = app.Command("send", "send a link to the queue.")
	sendLink = send.Arg("link", "Link to song.").Required().String()
	sendUser = send.Arg("user", "User id to send link under.").Required().Uint32()

	// "newRoom" subcommand
	newRoom  = app.Command("newRoom", "Creates a new room.")
	roomName = newRoom.Arg("name", "Name of the room.").Required().String()

	getRoom     = app.Command("getRoom", "Query for a room by name.")
	getRoomName = getRoom.Arg("name", "Name of the room.").Required().String()
)

/*
 * Connect to the remote server. Remember to close the returned connection when
 * done.
 */
func connectToRemote() (*grpc.ClientConn, bepb.YtbBackendClient) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.FailOnNonTempDialError(true))

	conn, err := grpc.Dial(*remoteHost+":"+*remotePort, opts...)
	if err != nil {
		fmt.Printf("failed to dial server: %v\n", err)
		os.Exit(1)
	}

	client := bepb.NewYtbBackendClient(conn)
	return conn, client
}

/*
 * Handler to send a song to the remote server
 */
func sendCommand(client bepb.YtbBackendClient) {
	link := *sendLink
	_, err := client.SendSong(context.Background(), &bepb.Submission{
		Link:   link,
		UserId: *sendUser,
	})
	if err != nil {
		fmt.Printf("failed to call SendSong: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(link)
}

/*
 * Handler to list the songs in the playlist
 */
func playlistCommand(client bepb.YtbBackendClient) {
	playlist, err := client.GetPlaylist(context.Background(), &cmpb.Empty{})
	if err != nil {
		fmt.Printf("failed to call GetPlaylist: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < len(playlist.Songs); i++ {
		fmt.Printf("%3d. { id: %2d, user: %2d, title: %s }\n",
			i+1, playlist.Songs[i].SongId, playlist.Songs[i].UserId, playlist.Songs[i].Title)
	}
}

/*
 * Tell the backend server to save the current playlist to a file
 */
func saveCommand(client bepb.YtbBackendClient) {
	response, err := client.SavePlaylist(context.Background(), &bepb.FilePath{Path: *saveFile})
	if err != nil {
		fmt.Printf("failed to call GetPlaylist: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: {success: %t, message: %s}\n", response.Success, response.Message)
}

func popCommand(client bepb.YtbBackendClient) {
	song, err := client.PopQueue(context.Background(), &cmpb.Empty{})
	if err != nil {
		fmt.Printf("failed to call PopQueue: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Popped song: { %v}\n", song)
}

func loginCommand(client bepb.YtbBackendClient) {
	user, err := client.LoginUser(context.Background(), &bepb.User{Username: *loginName, UserId: *loginId, RoomId: *loginRoomId})
	if err != nil {
		fmt.Printf("failed to call LoginUser: %v\n", err)
		os.Exit(1)
	}

	if user.Err.Success == false {
		fmt.Println(user.Err.Message)
	} else {
		fmt.Printf("User name: %s\n", user.Username)
		fmt.Printf("User id: %d\n", user.UserId)
		fmt.Printf("Room id: %d\n", user.RoomId)
	}
}

func removeCommand(client bepb.YtbBackendClient) {
	response, err := client.RemoveSong(context.Background(), &bepb.Eviction{SongId: *removeSong, UserId: *removeUser})
	if err != nil {
		fmt.Printf("failed to call RemoveSong: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: {success: %t, message: %s}\n", response.Success, response.Message)
}

func nowCommand(client bepb.YtbBackendClient) {
	song, err := client.GetNowPlaying(context.Background(), &cmpb.Empty{})
	if err != nil {
		fmt.Printf("failed to call GetNowPlaying: %v\n", err)
		os.Exit(1)
	}

	if song.SongId == 0 {
		fmt.Println("No song is currently playing")
	} else {
		fmt.Printf("Now Playing: { id: %2d, user: %2d, title: %s }\n",
			song.SongId, song.UserId, song.Title)
	}
}

func nextCommand(client bepb.YtbBackendClient) {
	response, err := client.NextSong(context.Background(), &cmpb.Empty{})
	if err != nil {
		fmt.Printf("failed to call NextSong: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: {success: %t, message: %s}\n", response.GetSuccess(), response.GetMessage())
}

func pauseCommand(client bepb.YtbBackendClient) {
	response, err := client.PauseSong(context.Background(), &cmpb.Empty{})
	if err != nil {
		fmt.Printf("failed to call PauseSong: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: {success: %t, message: %s}\n", response.GetSuccess(), response.GetMessage())
}

func newRoomCommand(client bepb.YtbBackendClient) {
	room, err := client.CreateRoom(context.Background(), &bepb.Room{Name: *roomName})
	if err != nil {
		fmt.Printf("failed to call CreateRoom: %v\n", err)
		os.Exit(1)
	}

	if room.Err.Success == false {
		fmt.Println(room.Err.Message)
	} else {
		fmt.Printf("Room name: %s\n", room.Name)
		fmt.Printf("Room id: %2d\n", room.Id)
	}
}

func getRoomCommand(client bepb.YtbBackendClient) {
	room, err := client.GetRoom(context.Background(), &bepb.Room{Name: *getRoomName})
	if err != nil {
		fmt.Printf("failed to call CreateRoom: %v\n", err)
		os.Exit(1)
	}

	if room.Err.Success == false {
		fmt.Println(room.Err.Message)
	} else {
		fmt.Printf("Room name: %s\n", room.Name)
		fmt.Printf("Room id: %2d\n", room.Id)
	}
}

func main() {
	kingpin.Version("0.1")
	parsed := kingpin.MustParse(app.Parse(os.Args[1:]))

	conn, client := connectToRemote()
	defer conn.Close()

	switch parsed {
	case now.FullCommand():
		nowCommand(client)

	case send.FullCommand():
		sendCommand(client)

	case playlist.FullCommand():
		playlistCommand(client)

	case save.FullCommand():
		saveCommand(client)

	case pop.FullCommand():
		popCommand(client)

	case login.FullCommand():
		loginCommand(client)

	case remove.FullCommand():
		removeCommand(client)

	case next.FullCommand():
		nextCommand(client)

	case pause.FullCommand():
		pauseCommand(client)

	case newRoom.FullCommand():
		newRoomCommand(client)

	case getRoom.FullCommand():
		getRoomCommand(client)

	default:
		nowCommand(client)
	}
}
