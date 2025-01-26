package server

import "github.com/gin-gonic/gin"

func (s *Server) registerHealthRoute() error {
	s.Engine.GET(s.Config.ApiBaseUrl+"/health", s.getHealthHandler())
	return nil
}

func (s *Server) getHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
}
