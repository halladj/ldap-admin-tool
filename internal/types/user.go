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

// UserDetails holds the attributes returned by a user query.
type UserDetails struct {
	DN        string
	UID       string
	FirstName string
	LastName  string
	Email     string
	UIDNumber int
	GIDNumber int
	HomeDir   string
	Shell     string
	Groups    []string
}

// GroupDetails holds the attributes returned by a group query.
type GroupDetails struct {
	DN      string
	Name    string
	GID     int
	Members []string
}
