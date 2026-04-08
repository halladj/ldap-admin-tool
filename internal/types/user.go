package types

// User represents an LDAP user account with both LDAP and PDF metadata.
type User struct {
	FirstName string
	LastName  string
	UID       string
	Email     string
	Password  string
	GID       int
	Groups    []string // list of groups the user belongs to
}
