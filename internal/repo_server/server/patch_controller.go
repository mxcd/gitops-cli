package server

import (
	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/patch"
)

func (s *Server) registerPatchRoute() error {
	s.Engine.PUT(s.Options.ApiBaseUrl+"/patch", s.getPatchHandler())
	return nil
}

func (s *Server) getPatchHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input patch.PatchTask
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		err := s.GitPatcher.Patch([]patch.PatchTask{input})
		if err != nil {
			c.JSON(500, gin.H{"error": "error executing patching"})
			return
		}

		c.JSON(200, gin.H{"message": "ok"})
	}
}