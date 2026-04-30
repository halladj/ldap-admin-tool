package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/halladj/ldap-admin-tool/internal/types"
)

// listGroups godoc
//
//	@Summary	List all groups
//	@Tags		groups
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Success	200	{array}		types.GroupDetails
//	@Failure	500	{object}	ErrorResponse
//	@Router		/groups [get]
func (s *Server) listGroups(c *gin.Context) {
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	groups, err := client.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// getGroup godoc
//
//	@Summary	Get a group by name
//	@Tags		groups
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		name	path		string	true	"Group name"
//	@Success	200		{object}	types.GroupDetails
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/groups/{name} [get]
func (s *Server) getGroup(c *gin.Context) {
	name := c.Param("name")
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	group, err := client.QueryGroup(name)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// createGroup godoc
//
//	@Summary	Create a new group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body		CreateGroupRequest	true	"Group creation request"
//	@Success	201		{object}	types.GroupDetails
//	@Failure	400		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/groups [post]
func (s *Server) createGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	gid := req.GID
	if gid == 0 {
		var err error
		gid, err = client.GetNextGIDNumber()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if err := client.CreateGroup(req.Name, gid); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, types.GroupDetails{Name: req.Name, GID: gid})
}

// deleteGroup godoc
//
//	@Summary	Delete a group
//	@Tags		groups
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		name	path	string	true	"Group name"
//	@Success	204
//	@Failure	404	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/groups/{name} [delete]
func (s *Server) deleteGroup(c *gin.Context) {
	name := c.Param("name")
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	if err := client.RemoveGroup(name); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// addGroupMembers godoc
//
//	@Summary	Add members to a group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		name	path	string				true	"Group name"
//	@Param		body	body	GroupMembersRequest	true	"UIDs to add"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/groups/{name}/members [post]
func (s *Server) addGroupMembers(c *gin.Context) {
	name := c.Param("name")
	var req GroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	for _, uid := range req.UIDs {
		if err := client.AddToGroup(uid, name); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.Status(http.StatusNoContent)
}

// removeGroupMembers godoc
//
//	@Summary	Remove members from a group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		name	path	string				true	"Group name"
//	@Param		body	body	GroupMembersRequest	true	"UIDs to remove"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/groups/{name}/members [delete]
func (s *Server) removeGroupMembers(c *gin.Context) {
	name := c.Param("name")
	var req GroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	for _, uid := range req.UIDs {
		if err := client.RemoveFromGroup(uid, name); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.Status(http.StatusNoContent)
}
