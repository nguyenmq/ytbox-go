/*
 * RPC client used by the frontend to interact with the backend song server.
 */

package frontend

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

type BackendClient struct {
	connection *grpc.ClientConn      // grpc connection
	be_client  bepb.YtbBackendClient // backend client
}

func (c *BackendClient) Connect(host string, port string) error {
	var err error
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.FailOnNonTempDialError(true))

	c.connection, err = grpc.Dial(host+":"+port, opts...)
	if err != nil {
		log.Fatalf("Failed to dial server %s on port %s with error: %v\n", host, port, err)
	} else {
		log.Println("Connected to backend server")
		c.be_client = bepb.NewYtbBackendClient(c.connection)
	}

	return err
}

func (c *BackendClient) GetPlaylist() (*bepb.Playlist, error) {
	playlist, err := c.be_client.GetPlaylist(context.Background(), &cmpb.Empty{})

	if err != nil {
		log.Printf("Failed to fetch playlist with error: %v\n", err)
	}

	return playlist, err
}

func (c *BackendClient) SendNewSong(link string, user_id uint32) (*bepb.Error, error) {
	var submission = bepb.Submission{
		Link:   link,
		UserId: user_id,
	}

	response, err := c.be_client.SendSong(context.Background(), &submission)

	if err != nil {
		log.Printf("Failed to send new song with error: %v\n", err)
	}

	return response, err
}

func (c *BackendClient) GetNowPlaying() (*cmpb.Song, error) {
	song, err := c.be_client.GetNowPlaying(context.Background(), &cmpb.Empty{})

	if err != nil {
		log.Printf("Failed to fetch currently playing song with error: %v\n", err)
	}

	return song, err
}
