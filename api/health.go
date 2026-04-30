package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status string `json:"status"`
	LDAP   string `json:"ldap"`
	Time   string `json:"time"`
}

func (s *Server) healthCheck(c *gin.Context) {
	resp := healthResponse{
		Time: time.Now().UTC().Format(time.RFC3339),
	}

	client, err := s.newClient()
	if err != nil {
		resp.Status = "degraded"
		resp.LDAP = "unreachable"
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}
	client.Close()

	resp.Status = "ok"
	resp.LDAP = "ok"
	c.JSON(http.StatusOK, resp)
}
