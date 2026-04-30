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
	DN        string   `json:"dn,omitempty"`
	UID       string   `json:"uid"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	UIDNumber int      `json:"uid_number"`
	GIDNumber int      `json:"gid_number"`
	HomeDir   string   `json:"home_dir"`
	Shell     string   `json:"shell"`
	Groups    []string `json:"groups,omitempty"`
}

// GroupDetails holds the attributes returned by a group query.
type GroupDetails struct {
	DN      string   `json:"dn,omitempty"`
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members,omitempty"`
}
