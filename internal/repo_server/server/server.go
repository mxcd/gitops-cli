package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/git"
)

type RouterConfig struct {
	DevMode    bool
	Port       int
	ApiBaseUrl string
	ApiKeys    []string
}

type Server struct {
	Engine     *gin.Engine
	HttpServer *http.Server
	Config     *RouterConfig
}

func NewServer(config *RouterConfig, gitConnection *git.Connection) (*Server, error) {
	if !config.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	router := &Server{
		Config: config,
		Engine: engine,
		HttpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: engine,
		},
	}

	return router, nil
}

func (s *Server) RegisterMiddlewares() {
	apiKeyAuthMiddleware := GetApiKeyAuthMiddleware(&ApiKeyAuthMiddlewareConfig{
		AllowedApiKeys: s.Config.ApiKeys,
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
