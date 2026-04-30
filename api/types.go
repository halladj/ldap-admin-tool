package api

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	FirstName string   `json:"first_name" binding:"required" example:"John"`
	LastName  string   `json:"last_name"  binding:"required" example:"Doe"`
	Email     string   `json:"email"      binding:"required" example:"jdoe@example.com"`
	UID       string   `json:"uid,omitempty"      example:"jdoe"`
	Password  string   `json:"password,omitempty" example:"MyP@ss123!"`
	GID       int      `json:"gid,omitempty"      example:"10008"`
	Groups    []string `json:"groups,omitempty"`
	SendEmail bool     `json:"send_email" example:"true"`
	SendPDF   bool     `json:"send_pdf"   example:"true"`
}

// CreateUserResponse is returned after successful user creation.
type CreateUserResponse struct {
	UID       string   `json:"uid"`
	UIDNumber int      `json:"uid_number"`
	Groups    []string `json:"groups,omitempty"`
	EmailSent bool     `json:"email_sent"`
}

// ChangePasswordRequest is the request body for changing a user's password.
// Leave password empty to auto-generate.
type ChangePasswordRequest struct {
	Password string `json:"password,omitempty" example:"NewP@ss123!"`
}

// ChangePasswordResponse is returned after a password change.
// The new password is included only when it was auto-generated.
type ChangePasswordResponse struct {
	Password string `json:"password,omitempty"`
}

// ChangeEmailRequest is the request body for changing a user's email.
type ChangeEmailRequest struct {
	Email string `json:"email" binding:"required" example:"newemail@example.com"`
}

// UserGroupsRequest is the request body for adding/removing a user to/from groups.
type UserGroupsRequest struct {
	Groups []string `json:"groups" binding:"required"`
}

// CreateGroupRequest is the request body for creating a group.
type CreateGroupRequest struct {
	Name string `json:"name" binding:"required" example:"printing-c"`
	GID  int    `json:"gid,omitempty"          example:"10050"`
}

// GroupMembersRequest is the request body for adding/removing members from a group.
type GroupMembersRequest struct {
	UIDs []string `json:"uids" binding:"required"`
}

// ErrorResponse is returned on any error.
type ErrorResponse struct {
	Error string `json:"error" example:"user 'jdoe' not found"`
}
