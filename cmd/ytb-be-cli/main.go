package main

import (
	"log"

	"github.com/nguyenmq/ytbox-go/common"
	pb "github.com/nguyenmq/ytbox-go/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	prefix string = "ytb-be-cli"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	logFile := ytb_common.InitLogger(prefix, true)
	defer logFile.Close()

	conn, err := grpc.Dial("127.0.0.1:8000", opts...)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewYtbBackendClient(conn)

	errorMsg, err := client.SubmitSong(context.Background(), &pb.UserSubmission{"youtube.com", 0})
	if err != nil {
		log.Fatalf("failed to call SubmitSong: %v", err)
	}

	log.Println(errorMsg.Errno)
	log.Println(errorMsg.Title)
	log.Println(errorMsg.Body)
}
