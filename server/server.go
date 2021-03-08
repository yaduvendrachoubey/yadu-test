package server

import "github.com/gin-gonic/gin"

type Server struct {
	GinEngine *gin.Engine
}

func NewServer() *Server {
	g := gin.New()
	return &Server{
		GinEngine: g,
	}

}

func (s *Server) Run() {
	s.GinEngine.Run()
}
