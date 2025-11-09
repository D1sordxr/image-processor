package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"

	"github.com/D1sordxr/image-processor/internal/transport/http/setup"

	"github.com/D1sordxr/image-processor/internal/domain/app/port"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/D1sordxr/image-processor/internal/transport/http/middleware"

	"github.com/wb-go/wbf/ginext"
)

type routeRegisterer interface {
	RegisterRoutes(router *ginext.RouterGroup)
}

type Server struct {
	log      port.Logger
	cfg      *config.HTTPServer
	engine   *ginext.Engine
	server   *http.Server
	handlers []routeRegisterer
}

func NewServer(
	log port.Logger,
	cfg *config.HTTPServer,
	handlers ...routeRegisterer,
) *Server {
	log.Info("Initializing HTTP server", "port", cfg.Port)

	var engine *ginext.Engine
	if cfg.ReleaseMode {
		engine = ginext.New(setup.ReleaseMode)
	} else {
		engine = ginext.New("")
	}
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())

	if cfg.CORS {
		allowedOrigins := cfg.AllowOrigins
		if len(allowedOrigins) == 0 {
			allowedOrigins = []string{"*"}
		}

		engine.Use(middleware.CORS(middleware.CORSConfig{
			AllowOrigins:     allowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	return &Server{
		log: log,
		server: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           engine.Handler(),
			ReadHeaderTimeout: cfg.Timeout,
			ReadTimeout:       cfg.Timeout,
			WriteTimeout:      cfg.Timeout,
		},
		engine:   engine,
		handlers: handlers,
	}
}

func (s *Server) Run(_ context.Context) error {
	if s.cfg.ServeUI { // TODO fix panic
		uiHandler := func(c *gin.Context) {
			if _, err := os.Stat(s.cfg.UIPath); os.IsNotExist(err) {
				s.log.Error("UI file not found", "path", s.cfg.UIPath)
				c.JSON(500, gin.H{"error": "UI not available"})
				return
			}
			c.File(s.cfg.UIPath)
		}
		s.log.Info("Serving UI",
			"path", fmt.Sprintf("http://%s:%s/ui/index.html", s.cfg.Host, s.cfg.Port),
		)

		// TODO s.engine.Static("/static", "ui/static")
		s.engine.GET("/", uiHandler)
		s.engine.NoRoute(uiHandler)
	}

	s.log.Info("Registering HTTP handlers...")
	for _, handler := range s.handlers {
		group := s.engine.Group("/api")
		handler.RegisterRoutes(group)
	}

	s.log.Info("Starting HTTP server...", "address", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.log.Info("HTTP server closed gracefully")
			return nil
		}
		s.log.Error("HTTP server stopped with error", "error", err.Error())
		return err
	}

	s.log.Info("HTTP server exited unexpectedly")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server...")
	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Error("Failed to gracefully shutdown HTTP server", "error", err.Error())
		return err
	}
	s.log.Info("HTTP server shutdown complete")
	return nil
}
