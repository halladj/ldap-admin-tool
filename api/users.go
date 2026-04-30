package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/halladj/ldap-admin-tool/internal/mail"
	"github.com/halladj/ldap-admin-tool/internal/password"
	"github.com/halladj/ldap-admin-tool/internal/pdf"
	"github.com/halladj/ldap-admin-tool/internal/types"
)

// listUsers godoc
//
//	@Summary	List all users
//	@Tags		users
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Success	200	{array}		types.UserDetails
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users [get]
func (s *Server) listUsers(c *gin.Context) {
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	users, err := client.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// getUser godoc
//
//	@Summary	Get a user by UID
//	@Tags		users
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid	path		string	true	"Username"
//	@Success	200	{object}	types.UserDetails
//	@Failure	404	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users/{uid} [get]
func (s *Server) getUser(c *gin.Context) {
	uid := c.Param("uid")
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	user, err := client.QueryUser(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// createUser godoc
//
//	@Summary	Create a new user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body		CreateUserRequest	true	"User creation request"
//	@Success	201		{object}	CreateUserResponse
//	@Failure	400		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/users [post]
func (s *Server) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.UID == "" {
		req.UID = strings.ToLower(string(req.FirstName[0]) + req.LastName)
	}
	if req.Password == "" {
		var err error
		req.Password, err = password.Generate(12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("failed to generate password: %v", err)})
			return
		}
	}
	if req.GID == 0 {
		req.GID = s.cfg.DefaultGID
	}

	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	uidNumber, err := client.GetNextUIDNumber()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	user := types.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UID:       req.UID,
		Email:     req.Email,
		Password:  req.Password,
		GID:       req.GID,
	}

	if _, err := client.CreateUser(user, uidNumber); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	for _, g := range req.Groups {
		_ = client.AddToGroup(req.UID, g)
	}
	user.Groups = req.Groups

	resp := CreateUserResponse{
		UID:       req.UID,
		UIDNumber: uidNumber,
		Groups:    req.Groups,
	}

	if req.SendPDF || req.SendEmail {
		pdfBytes, err := pdf.Generate(s.cfg, user)
		if err == nil && req.SendEmail {
			if err := mail.SendWelcome(mail.WelcomeEmail{
				From:      s.cfg.SenderEmail,
				To:        req.Email,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				UID:       req.UID,
				PDF:       pdfBytes,
			}); err == nil {
				resp.EmailSent = true
			}
		}
	}

	c.JSON(http.StatusCreated, resp)
}

// deleteUser godoc
//
//	@Summary	Delete a user
//	@Tags		users
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid	path	string	true	"Username"
//	@Success	204
//	@Failure	404	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users/{uid} [delete]
func (s *Server) deleteUser(c *gin.Context) {
	uid := c.Param("uid")
	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	if err := client.DeleteUser(uid); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// changePassword godoc
//
//	@Summary	Change a user's password
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid		path		string					true	"Username"
//	@Param		body	body		ChangePasswordRequest	false	"Leave password empty to auto-generate"
//	@Success	200		{object}	ChangePasswordResponse
//	@Failure	400		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/users/{uid}/password [put]
func (s *Server) changePassword(c *gin.Context) {
	uid := c.Param("uid")
	var req ChangePasswordRequest
	_ = c.ShouldBindJSON(&req)

	newPass := req.Password
	if newPass == "" {
		var err error
		newPass, err = password.Generate(12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}

	client, err := s.newClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer client.Close()

	if err := client.ChangePassword(uid, newPass); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := ChangePasswordResponse{}
	if req.Password == "" {
		resp.Password = newPass
	}
	c.JSON(http.StatusOK, resp)
}

// changeEmail godoc
//
//	@Summary	Change a user's email
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid		path	string				true	"Username"
//	@Param		body	body	ChangeEmailRequest	true	"New email address"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users/{uid}/email [put]
func (s *Server) changeEmail(c *gin.Context) {
	uid := c.Param("uid")
	var req ChangeEmailRequest
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

	if err := client.ChangeEmail(uid, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// addUserToGroups godoc
//
//	@Summary	Add a user to groups
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid		path	string				true	"Username"
//	@Param		body	body	UserGroupsRequest	true	"Groups to add"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users/{uid}/groups [post]
func (s *Server) addUserToGroups(c *gin.Context) {
	uid := c.Param("uid")
	var req UserGroupsRequest
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

	for _, g := range req.Groups {
		if err := client.AddToGroup(uid, g); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.Status(http.StatusNoContent)
}

// removeUserFromGroups godoc
//
//	@Summary	Remove a user from groups
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		uid		path	string				true	"Username"
//	@Param		body	body	UserGroupsRequest	true	"Groups to remove from"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users/{uid}/groups [delete]
func (s *Server) removeUserFromGroups(c *gin.Context) {
	uid := c.Param("uid")
	var req UserGroupsRequest
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

	for _, g := range req.Groups {
		if err := client.RemoveFromGroup(uid, g); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}
	c.Status(http.StatusNoContent)
}
