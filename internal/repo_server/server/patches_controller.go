package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/patch"
	"github.com/rs/zerolog/log"
)

func (s *Server) registerPatchesRoute() error {
	s.Engine.PUT(s.Options.ApiBaseUrl+"/patches", s.getPatchesHandler())
	return nil
}

func (s *Server) getPatchesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input patch.PatchBatch
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		// validate before acquiring the lock so invalid requests do not block
		// concurrent patches
		if err := patch.ValidatePatchBatch(input); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		s.patchLock.Lock()
		defer s.patchLock.Unlock()

		commitHash, err := s.GitPatcher.PatchBatch(input)
		if err != nil {
			if errors.Is(err, patch.ErrInvalidPatchBatch) {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			log.Error().Err(err).Msg("Error executing batch patching")
			c.JSON(500, gin.H{"error": "error executing patching"})
			return
		}

		c.JSON(200, gin.H{"message": "ok", "commit": commitHash})
	}
}
