/*
 * Implements the web front for users to submit and view songs in yt_box.
 */

package frontend

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	LogPrefix string = "ytb-fe" // logging prefix name
)

type FrontendServer struct {
	addr   string       // ip address and port to listen on
	router *gin.Engine  // gin router
	server *http.Server // http server
}

func NewServer(addr string) *FrontendServer {
	frontend := new(FrontendServer)
	frontend.addr = addr
	frontend.router = gin.Default()

	frontend.server = new(http.Server)
	frontend.server.Addr = frontend.addr
	frontend.server.Handler = frontend.router

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
