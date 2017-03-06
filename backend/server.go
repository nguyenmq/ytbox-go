// Implements the RPC server for yt_box backend

package ytbbe

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/nguyenmq/ytbox-go/proto"
)

const (
	LogPrefix string = "ytb-be" // logging prefix name
)

/*
 * Implements the backend rpc server interface
 */
type YtbBackendServer struct {
	listener   net.Listener // network listener
	grpcServer *grpc.Server // grpc server
}

/*
 * Create a new yt_box backend server
 */
func NewServer(addr string) *YtbBackendServer {
	var err error
	server := new(YtbBackendServer)

	server.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s with error: %v", addr, err)
	}

	server.grpcServer = grpc.NewServer()
	pb.RegisterYtbBackendServer(server.grpcServer, server)

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
func (s *YtbBackendServer) SubmitSong(con context.Context, sub *pb.UserSubmission) (*pb.ErrorMessage, error) {
	log.Printf("link: %s\n", sub.Link)
	log.Printf("user id: %d\n", sub.UserId)

	return &pb.ErrorMessage{
		Errno: 0,
		Title: "Success",
		Body:  ""}, nil
}
