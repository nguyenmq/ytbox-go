/*
 * RPC client used by the frontend to interact with the backend song server.
 */

package frontend

import (
	"errors"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	bepb "github.com/nguyenmq/ytbox-go/internal/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
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

	if !response.Success {
		err = errors.New(response.Message)
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

func (c *BackendClient) RemoveSong(song_id uint32, user_id uint32) (*bepb.Error, error) {
	var eviction_request = bepb.Eviction{
		SongId: song_id,
		UserId: user_id,
	}

	response, err := c.be_client.RemoveSong(context.Background(), &eviction_request)

	if err != nil {
		log.Printf("Failed to remove song with error: %v\n", err)
	}

	return response, err
}

func (c *BackendClient) LoginNewUser(userName string, roomName string) (*bepb.User, error) {
	roomRequest := bepb.Room{Name: roomName}

	room, err := c.be_client.GetRoom(context.Background(), &roomRequest)
	if err != nil {
		log.Printf("Failed to describe room with error: %v\n", err)
		return nil, err
	}
	if !room.Err.Success {
		log.Printf("Requested room %s does not exist\n", roomName)
		return nil, ErrRoomNotFound
	}

	userRequest := bepb.User{Username: userName, RoomId: room.Id}
	user, err := c.be_client.LoginUser(context.Background(), &userRequest)
	if err != nil {
		log.Printf("Failed to login user with error: %v\n", err)
		return nil, err
	}

	if user.UserId == 0 {
		log.Printf("Failed to login")
		return nil, ErrFailedLogin
	}

	return user, err
}

func (c *BackendClient) NextSong() (*bepb.Error, error) {
	response, err := c.be_client.NextSong(context.Background(), &cmpb.Empty{})

	if err != nil {
		log.Printf("Failed to skip currently playing song with error: %v\n", err)
	}

	return response, err
}
