/*
 * Implements the RPC server for yt_box backend
 */

package backend

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	sched "github.com/nguyenmq/ytbox-go/backend/scheduler"
	pb "github.com/nguyenmq/ytbox-go/proto/backend"
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
func NewServer(addr string, loadFile string) *YtbBackendServer {
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

	// load a snapshot playlist if provided
	if loadFile != "" {
		server.loadPlaylistFromFile(loadFile)
	}

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
func (s *YtbBackendServer) SubmitSong(con context.Context, sub *pb.Submission) (*pb.Error, error) {
	var response *pb.Error = new(pb.Error)
	log.Printf("Submission: {link: %s, userId: %d}\n", sub.Link, sub.UserId)

	song, err := fetchSongData(sub.Link, sub.UserId)
	log.Printf("Song data: { %v}", song)

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

/*
 * Load a playlist from a serialized protobuf file
 */
func (s *YtbBackendServer) loadPlaylistFromFile(file string) {
	var in []byte
	var err error

	in, err = ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading file: %s", file)
		return
	}

	playlist := &pb.Playlist{}
	err = proto.Unmarshal(in, playlist)
	if err != nil {
		log.Printf("Failed to parse playlist file: %v", err)
		return
	}

	log.Printf("Loading songs from file \"%s\":", file)
	for i := 0; i < len(playlist.Songs); i++ {
		song := &pb.Song{
			Title:     playlist.Songs[i].Title,
			SongId:    playlist.Songs[i].SongId,
			Username:  playlist.Songs[i].Username,
			UserId:    playlist.Songs[i].UserId,
			Service:   playlist.Songs[i].Service,
			ServiceId: playlist.Songs[i].ServiceId,
		}
		s.queue.AddSong(song)
		log.Printf("%3d. %s", i+1, song.Title)
	}
}

/*
 * Returns the songs in the queue back to the requesting client
 */
func (s *YtbBackendServer) GetPlaylist(con context.Context, arg *pb.Empty) (*pb.Playlist, error) {
	return s.queue.GetPlaylist(), nil
}

/*
 * Saves the current playlist to the given file location
 */
func (s *YtbBackendServer) SavePlaylist(con context.Context, fname *pb.FilePath) (*pb.Error, error) {
	response := &pb.Error{Success: false}
	playlist := s.queue.GetPlaylist()

	out, err := proto.Marshal(playlist)
	if err != nil {
		log.Printf("Failed to encode Playlist with error: %v", err)
		response.Message = err.Error()
		return response, nil
	}

	err = ioutil.WriteFile(fname.Path, out, 0644)
	if err != nil {
		log.Printf("Failed to write playlist to file \"%s\" with error: %v", fname.Path, err)
		response.Message = err.Error()
		return response, nil
	}

	log.Printf("Saved current playlist to: %s", fname.Path)
	response.Message = "Success"
	return response, nil
}
