/*
 * Implements the web front for users to submit and view songs in yt_box.
 */

package frontend

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	LogPrefix      string = "ytb-fe" // logging prefix name
	titleMaxLength int    = 100      // the maximum length of the now playing title
)

type FrontendServer struct {
	addr   string         // ip address and port to listen on
	client *BackendClient // the backend client
	router *gin.Engine    // gin router
	server *http.Server   // http server
	store  cookie.Store   // session cookie store
}

func NewServer(addr string) *FrontendServer {
	frontend := new(FrontendServer)
	frontend.addr = addr
	frontend.store = cookie.NewStore([]byte("tmp_secret"))
	frontend.store.Options(sessions.Options{
		MaxAge: 0,
	})

	// set up gin router
	frontend.router = gin.Default()
	frontend.router.HTMLRender = ginview.Default()
	frontend.router.Static("/static", "./static")
	frontend.router.StaticFile("/favicon.ico", "./static/img/favicon.ico")
	frontend.router.Use(sessions.Sessions("yt_box", frontend.store))

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
	frontend.router.GET("/now_playing", frontend.HandleNowPlaying)
	frontend.router.POST("/remove", frontend.HandleRemove)
	frontend.router.GET("/login", frontend.HandleLoginPage)
	frontend.router.POST("/login", frontend.HandleLoginPost)
	frontend.router.GET("/next", frontend.HandleNextSong)
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
	session := sessions.Default(context)

	if session.Get("user_id") == nil {
		context.Redirect(http.StatusTemporaryRedirect, "/login")
	} else {
		title := "No song is currently playing"

		current_song, err := s.client.GetNowPlaying()
		has_song_playing := current_song.SongId != 0

		if err == nil && has_song_playing {
			title = truncate_song_title(current_song.Title, titleMaxLength)
		}

		playlist, err := s.client.GetPlaylist()

		user_id := session.Get("user_id").(uint32)

		context.HTML(http.StatusOK, "index", gin.H{
			"title":                "yt-box: Song Queue",
			"now_playing":          title,
			"has_song_playing":     has_song_playing,
			"song":                 current_song,
			"song_count":           len(playlist.Songs),
			"queue":                playlist.Songs,
			"session_user_id":      user_id,
			"increment_index":      increment_index,
			"transform_user_name":  s.TransformUsername,
			"matches_session_user": s.MatchesSessionUser,
		})
	}
}

func (s *FrontendServer) HandlePlaylist(context *gin.Context) {
	playlist, err := s.client.GetPlaylist()

	session := sessions.Default(context)
	user_id := session.Get("user_id").(uint32)

	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to retrieve playlist")
	} else {
		context.HTML(http.StatusOK, "layouts/queue.html", gin.H{
			"song_count":           len(playlist.Songs),
			"queue":                playlist.Songs,
			"session_user_id":      user_id,
			"increment_index":      increment_index,
			"transform_user_name":  s.TransformUsername,
			"matches_session_user": s.MatchesSessionUser,
		})
	}
}

func (s *FrontendServer) HandleNewSong(context *gin.Context) {
	link, exists := context.GetPostForm("submit_box")

	if exists == false {
		context.String(http.StatusInternalServerError, "Failed to submit song")
		return
	}

	session := sessions.Default(context)
	user_id := session.Get("user_id").(uint32)

	_, err := s.client.SendNewSong(link, user_id)
	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to submit song")
	} else {
		context.String(http.StatusOK, "Success")
	}
}

func (s *FrontendServer) HandleNowPlaying(context *gin.Context) {
	title := "No song is currently playing"

	current_song, err := s.client.GetNowPlaying()
	has_song_playing := current_song.SongId != 0

	if err == nil && has_song_playing {
		title = truncate_song_title(current_song.Title, titleMaxLength)
	}

	session := sessions.Default(context)
	user_id := session.Get("user_id").(uint32)

	context.HTML(http.StatusOK, "layouts/now_playing.html", gin.H{
		"now_playing":          title,
		"has_song_playing":     has_song_playing,
		"song":                 current_song,
		"session_user_id":      user_id,
		"transform_user_name":  s.TransformUsername,
		"matches_session_user": s.MatchesSessionUser,
	})
}

func (s *FrontendServer) HandleRemove(context *gin.Context) {
	// todo: get user id from session
	song_id_str, exists := context.GetPostForm("song_id")

	if exists == false {
		context.String(http.StatusBadRequest, "Did not get a song to delete")
		return
	}

	song_id, err := strconv.ParseUint(song_id_str, 10, 32)
	session := sessions.Default(context)
	user_id := session.Get("user_id").(uint32)

	_, err = s.client.RemoveSong(uint32(song_id), user_id)
	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to remove song")
	} else {
		context.String(http.StatusOK, "Success")
	}
}

func (s *FrontendServer) HandleLoginPage(context *gin.Context) {
	// todo: check for cookie and redirect if already have cookie
	context.HTML(http.StatusOK, "login", gin.H{
		"title": "yt-box: Login",
	})
}

func (s *FrontendServer) HandleLoginPost(context *gin.Context) {
	user_name, exists := context.GetPostForm("submit_box")

	if exists == false {
		context.String(http.StatusInternalServerError, "Failed to get a valid login request")
		return
	}

	user, err := s.client.LoginNewUser(user_name)

	if err != nil {
		context.String(http.StatusInternalServerError, "Failed to login user")
		return
	}

	session := sessions.Default(context)
	session.Set("user_id", user.UserId)
	session.Save()
	context.Request.Method = "GET"
	context.Request.URL.Path = "/"
	s.router.HandleContext(context)
}

func (s *FrontendServer) HandleNextSong(context *gin.Context) {
	current_song, _ := s.client.GetNowPlaying()
	session := sessions.Default(context)
	user_id := session.Get("user_id").(uint32)

	if s.MatchesSessionUser(current_song.UserId, user_id) {
		s.client.NextSong()
	}

	context.Status(http.StatusOK)
}

func (s *FrontendServer) TransformUsername(song *cmpb.Song, session_user_id uint32) string {
	if song.UserId == session_user_id {
		return "You"
	} else {
		return song.Username
	}
}

func (s *FrontendServer) MatchesSessionUser(user_id uint32, session_user_id uint32) bool {
	return user_id == session_user_id
}

func increment_index(index int) int {
	return index + 1
}

func truncate_song_title(title string, length int) string {
	if len(title) > length {
		return fmt.Sprintf("%s…", title[0:length])
	}

	return title
}
