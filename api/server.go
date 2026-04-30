package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	_ "github.com/halladj/ldap-admin-tool/docs"
	"github.com/halladj/ldap-admin-tool/internal/config"
	ldapclient "github.com/halladj/ldap-admin-tool/internal/ldap"
)

// Server holds the config and admin credentials used for per-request LDAP connections.
type Server struct {
	cfg       *config.Config
	adminPass string
	log       *slog.Logger
}

// NewServer creates a new API server instance.
func NewServer(cfg *config.Config, adminPass string) *Server {
	return &Server{
		cfg:       cfg,
		adminPass: adminPass,
		log:       slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// Start registers all routes and begins listening on the given port.
// Blocks until SIGTERM or SIGINT, then drains in-flight requests for up to 10 seconds.
func (s *Server) Start(port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestID())
	r.Use(securityHeaders())
	r.Use(structuredLogger(s.log))
	r.Use(rateLimiter())

	// Public endpoints (no API key required)
	r.GET("/health", s.healthCheck)
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")
	v1.Use(apiKeyAuth(s.cfg.APIKey))

	v1.GET("/users", s.listUsers)
	v1.POST("/users", s.createUser)
	v1.GET("/users/:uid", s.getUser)
	v1.DELETE("/users/:uid", s.deleteUser)
	v1.PUT("/users/:uid/password", s.changePassword)
	v1.PUT("/users/:uid/email", s.changeEmail)
	v1.POST("/users/:uid/groups", s.addUserToGroups)
	v1.DELETE("/users/:uid/groups", s.removeUserFromGroups)

	v1.GET("/groups", s.listGroups)
	v1.POST("/groups", s.createGroup)
	v1.GET("/groups/:name", s.getGroup)
	v1.DELETE("/groups/:name", s.deleteGroup)
	v1.POST("/groups/:name/members", s.addGroupMembers)
	v1.DELETE("/groups/:name/members", s.removeGroupMembers)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("API server starting",
		"port", port,
		"swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", port),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	s.log.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func (s *Server) newClient() (*ldapclient.Client, error) {
	return ldapclient.NewClient(s.cfg, s.adminPass)
}
