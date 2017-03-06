// Entry point into the yt_box backend service

package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nguyenmq/ytbox-go/common"
	pb "github.com/nguyenmq/ytbox-go/proto"
)

const (
	// logging prefix name
	prefix string = "ytb-be"
)

type ytbBackendServer struct {
}

/*
 * Implements RPC function for a client to submit a song to the backend for
 * appending to the play queue.
 */
func (s *ytbBackendServer) SubmitSong(con context.Context, sub *pb.UserSubmission) (*pb.ErrorMessage, error) {
	log.Printf("link: %s\n", sub.Link)
	log.Printf("user id: %d\n", sub.UserId)

	return &pb.ErrorMessage{
		Errno: 0,
		Title: "Success",
		Body:  ""}, nil
}

func main() {
	logFile := ytb_common.InitLogger(prefix, true)
	defer logFile.Close()

	ytbServer := new(ytbBackendServer)

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	pb.RegisterYtbBackendServer(grpcServer, ytbServer)
	grpcServer.Serve(listener)
}
