/*
 * Implements the web front for users to submit and view songs in yt_box.
 */

package frontend

import (
	"log"
	"net/http"
	"os"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
)

const (
	LogPrefix string = "ytb-fe" // logging prefix name
)

type FrontendServer struct {
	addr   string         // ip address and port to listen on
	router *gin.Engine    // gin router
	server *http.Server   // http server
	client *BackendClient // the backend client
}

func NewServer(addr string) *FrontendServer {
	frontend := new(FrontendServer)
	frontend.addr = addr

	// set up gin router
	frontend.router = gin.Default()
	frontend.router.HTMLRender = ginview.Default()
	frontend.router.Static("/static", "./static")

	// set up the http server
	frontend.server = new(http.Server)
	frontend.server.Addr = frontend.addr
	frontend.server.Handler = frontend.router

	// connect to the song queue backend
	frontend.client = new(BackendClient)
	if err := frontend.client.Connect("127.0.0.1", "9009"); err != nil {
		os.Exit(1)
	}

	// configure routes
	frontend.router.GET("/", frontend.HandleIndex)
	frontend.router.GET("/playlist", frontend.HandlePlaylist)
	frontend.router.POST("/new_song", frontend.HandleNewSong)
	frontend.router.GET("/ping", func(context *gin.Context) {
		context.String(http.StatusOK, "pong")
	})

	return frontend
}

func (s *FrontendServer) Start() {
	if err := s.server.ListenAndServe(); err != nil {
		log.Println("Server closed by request")
	} else {
		log.Fatal("Server closed unexpectedly")
	}
}

func (s *FrontendServer) Stop() {
	s.server.Close()
}

func (s *FrontendServer) HandleIndex(context *gin.Context) {
	title := "No Song is currently playing"

	song, err := s.client.GetNowPlaying()

	if err == nil && song.SongId != 0 {
		title = song.Title
	}

	context.HTML(http.StatusOK, "index", gin.H{
		"title":       "yt-box Song Queue",
		"now_playing": title,
	})
}

func (s *FrontendServer) HandlePlaylist(context *gin.Context) {
	playlist, err := s.client.GetPlaylist()

	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to retrieve playlist")
	} else {
		for i := 0; i < len(playlist.Songs); i++ {
			song := playlist.Songs[i]
			context.String(http.StatusOK, "%3d. { id: %2d, user: %2d, title: %s }\n",
				i+1, song.SongId, song.UserId, song.Title)
		}
	}
}

func (s *FrontendServer) HandleNewSong(context *gin.Context) {
	link, exists := context.GetPostForm("submit_box")

	if exists == false {
		context.String(http.StatusInternalServerError, "Failed to submit song")
		return
	}

	_, err := s.client.SendNewSong(link, 7)
	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to submit song")
	} else {
		context.String(http.StatusOK, "Success")
	}
}
