package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/patch"
)

type RouterOptions struct {
	DevMode    bool
	Port       int
	ApiBaseUrl string
	ApiKeys    []string
}

type Server struct {
	Engine     *gin.Engine
	HttpServer *http.Server
	Options    *RouterOptions
	GitPatcher *patch.GitPatcher
}

func NewServer(options *RouterOptions, gitPatcher *patch.GitPatcher) (*Server, error) {
	if !options.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	router := &Server{
		Options: options,
		Engine:  engine,
		HttpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: engine,
		},
		GitPatcher: gitPatcher,
	}

	return router, nil
}

func (s *Server) RegisterMiddlewares() {
	apiKeyAuthMiddleware := GetApiKeyAuthMiddleware(&ApiKeyAuthMiddlewareConfig{
		AllowedApiKeys: s.Options.ApiKeys,
		UnprotectedRoutes: []string{
			s.Options.ApiBaseUrl + "/health",
		},
	})
	s.Engine.Use(apiKeyAuthMiddleware)
}

func (s *Server) RegisterRoutes() error {
	s.registerHealthRoute()
	s.registerPatchRoute()

	return nil
}

func (s *Server) Run() error {
	if err := s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) {
	s.HttpServer.Shutdown(ctx)
}
