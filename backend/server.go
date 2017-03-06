/*
 *Implements the RPC server for yt_box backend
 */

package backend

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	sched "github.com/nguyenmq/ytbox-go/backend/scheduler"
	pb "github.com/nguyenmq/ytbox-go/proto"
)

const (
	LogPrefix string = "ytb-be" // logging prefix name
)

/*
 * Implements the backend rpc server interface
 */
type YtbBackendServer struct {
	listener   net.Listener         // network listener
	grpcServer *grpc.Server         // grpc server
	queue      sched.QueueScheduler // playlist queue
}

/*
 * Create a new yt_box backend server
 */
func NewServer(addr string) *YtbBackendServer {
	var err error

	// initialize the backend server struct
	server := new(YtbBackendServer)
	server.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s with error: %v", addr, err)
	}

	// initialize the rpc server
	server.grpcServer = grpc.NewServer()
	pb.RegisterYtbBackendServer(server.grpcServer, server)

	// initialize the song queue
	server.queue = new(sched.FifoQueue)
	server.queue.Init()

	return server
}

/*
 * Start the server
 */
func (s *YtbBackendServer) Serve() {
	s.grpcServer.Serve(s.listener)
}

/*
 * Receive a song from a remote client for appending to the play queue
 */
func (s *YtbBackendServer) SubmitSong(con context.Context, sub *pb.SongSubmission) (*pb.ErrorMessage, error) {
	var response *pb.ErrorMessage = new(pb.ErrorMessage)
	log.Printf("Submission: {link: %s, userId: %d}\n", sub.Link, sub.UserId)

	song, err := fetchSongData(sub.Link, sub.UserId)
	log.Printf("Song data: %v", song)

	if err != nil {
		response.Success = false
		response.Message = err.Error()
	} else {
		response.Success = true
		response.Message = "Success"
		s.queue.AddSong(song)
	}

	return response, nil
}
