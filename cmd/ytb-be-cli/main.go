package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nguyenmq/ytbox-go/common"
	pb "github.com/nguyenmq/ytbox-go/proto"
)

const (
	prefix string = "ytb-be-cli"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	logFile := common.InitLogger(prefix, true)
	defer logFile.Close()

	conn, err := grpc.Dial("127.0.0.1:8000", opts...)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewYtbBackendClient(conn)

	response, err := client.SubmitSong(context.Background(), &pb.SongSubmission{"https://www.youtube.com/watch?v=cdIBxhONpC0", 0})
	if err != nil {
		log.Fatalf("failed to call SubmitSong: %v", err)
	}

	log.Printf("Response: {flag: %t, message: %s}", response.Success, response.Message)
}
