package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/patch"
	"github.com/rs/zerolog/log"
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

		s.patchLock.Lock()
		defer s.patchLock.Unlock()

		err := s.GitPatcher.Patch([]patch.PatchTask{input})
		if err != nil {
			if errors.Is(err, patch.ErrInvalidPatchBatch) {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			log.Error().Err(err).Msg("Error executing patching")
			c.JSON(500, gin.H{"error": "error executing patching"})
			return
		}

		c.JSON(200, gin.H{"message": "ok"})
	}
}
