package api

import (
	"fmt"

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
}

// NewServer creates a new API server instance.
func NewServer(cfg *config.Config, adminPass string) *Server {
	return &Server{cfg: cfg, adminPass: adminPass}
}

// Start registers all routes and begins listening on the given port.
func (s *Server) Start(port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	fmt.Printf("[*] API listening on :%d\n", port)
	fmt.Printf("[*] Swagger UI: http://localhost:%d/swagger/index.html\n", port)
	return r.Run(fmt.Sprintf(":%d", port))
}

func (s *Server) newClient() (*ldapclient.Client, error) {
	return ldapclient.NewClient(s.cfg, s.adminPass)
}
