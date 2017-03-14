package main

import (
	"fmt"
	"os"
	//"os/exec"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	//cmpb "github.com/nguyenmq/ytbox-go/proto/common"
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

// Pops the next song off the queue and plays it
//func PlayQueue(queue sched.QueueScheduler, play bool, stop <-chan int) {
//	var ticker <-chan time.Time = time.Tick(2 * time.Second)
//	var halt bool = false
//
//	for !halt {
//		select {
//		case <-ticker:
//			nextSong := queue.PopQueue()
//
//			if nextSong != nil {
//				fmt.Println("Popped song:", nextSong.Title)
//
//				if play {
//					link := fmt.Sprintf("https://www.youtube.com/watch?v=%s", nextSong.ServiceId)
//					err := exec.Command("mpv", "--fs", link).Run()
//
//					if err != nil {
//						fmt.Println("Failed to play link:", link)
//					}
//				}
//			}
//
//		case <-stop:
//			fmt.Println("Told to stop")
//			halt = true
//		}
//	}
//}

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

func main() {
	//var link string
	kingpin.Version("0.1")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	conn, client := connectToRemote()
	defer conn.Close()

	stream, err := client.SongPlayer(context.Background())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}

	stream.Send(&bepb.PlayerStatus{Command: bepb.CommandType_Ready})

	//song, err := client.PopQueue(context.Background(), &bepb.Empty{})
	//if err != nil {
	//	fmt.Printf("failed to call PopQueue: %v\n", err)
	//	os.Exit(1)
	//}

	//switch song.Service {
	//case cmpb.ServiceType_ServiceLocal:
	//	link = song.ServiceId

	//case cmpb.ServiceType_ServiceYoutube:
	//	link = fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ServiceId)

	//default:
	//	fmt.Printf("Unsupported link: %s\n", song.ServiceId)
	//}

	//if link != "" {
	//	err = exec.Command("mpv", "--fs", link).Run()
	//	if err != nil {
	//		fmt.Printf("Failed to play link: %s\n", link)
	//	}
	//}
}
