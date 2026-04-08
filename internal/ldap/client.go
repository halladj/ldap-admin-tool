package ldap

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-ldap/ldap/v3"
	"github.com/halladj/ldap-admin-tool/internal/config"
	"github.com/halladj/ldap-admin-tool/internal/types"
)

type Client struct {
	conn *ldap.Conn
	cfg  *config.Config
}

func NewClient(cfg *config.Config, adminPass string) (*Client, error) {
	parsed, _ := url.Parse(cfg.LDAPServer)
	serverName := parsed.Hostname()

	conn, err := ldap.DialURL(cfg.LDAPServer, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: false,
		ServerName:         serverName,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	if err := conn.Bind(cfg.AdminDN, adminPass); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind as admin: %w", err)
	}

	return &Client{conn: conn, cfg: cfg}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

// nextIDNumber finds the highest existing ID number in a search result and returns next available ID.
func (c *Client) nextIDNumber(searchBase, filter, attr string, floor int) (int, error) {
	req := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{attr},
		nil,
	)

	result, err := c.conn.Search(req)
	if err != nil {
		return 0, fmt.Errorf("failed to search %s: %w", searchBase, err)
	}

	max := floor
	for _, entry := range result.Entries {
		if n, err := strconv.Atoi(entry.GetAttributeValue(attr)); err == nil && n > max {
			max = n
		}
	}

	return max + 1, nil
}

func (c *Client) GetNextUIDNumber() (int, error) {
	return c.nextIDNumber(c.cfg.PeopleOU, "(objectClass=posixAccount)", "uidNumber", c.cfg.MinUIDNumber)
}

func (c *Client) CreateUser(user types.User, uidNumber int) (string, error) {
	cn := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	dn := fmt.Sprintf("cn=%s,%s", cn, c.cfg.PeopleOU)
	homeDir := fmt.Sprintf("/home/%s", user.UID)

	addReq := ldap.NewAddRequest(dn, nil)
	addReq.Attribute("objectClass", []string{"inetOrgPerson", "posixAccount", "top"})
	addReq.Attribute("cn", []string{cn})
	addReq.Attribute("sn", []string{user.LastName})
	addReq.Attribute("givenName", []string{user.FirstName})
	addReq.Attribute("uid", []string{user.UID})
	addReq.Attribute("uidNumber", []string{strconv.Itoa(uidNumber)})
	addReq.Attribute("gidNumber", []string{strconv.Itoa(user.GID)})
	addReq.Attribute("homeDirectory", []string{homeDir})
	addReq.Attribute("loginShell", []string{c.cfg.DefaultShell})
	addReq.Attribute("userPassword", []string{user.Password})
	addReq.Attribute("mail", []string{user.Email})

	if err := c.conn.Add(addReq); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return dn, nil
}

func (c *Client) AddToGroup(uid, groupName string) error {
	groupDN := fmt.Sprintf("cn=%s,%s", groupName, c.cfg.GroupOU)

	searchReq := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=posixGroup)",
		[]string{"memberUid"},
		nil,
	)

	result, err := c.conn.Search(searchReq)
	if err != nil {
		return fmt.Errorf("group '%s' not found: %w", groupName, err)
	}

	if len(result.Entries) > 0 {
		for _, member := range result.Entries[0].GetAttributeValues("memberUid") {
			if member == uid {
				return nil
			}
		}
	}

	modReq := ldap.NewModifyRequest(groupDN, nil)
	modReq.Add("memberUid", []string{uid})

	if err := c.conn.Modify(modReq); err != nil {
		return fmt.Errorf("failed to add to group '%s': %w", groupName, err)
	}

	return nil
}

func (c *Client) RemoveFromGroup(uid, groupName string) error {
	groupDN := fmt.Sprintf("cn=%s,%s", groupName, c.cfg.GroupOU)

	modReq := ldap.NewModifyRequest(groupDN, nil)
	modReq.Delete("memberUid", []string{uid})

	if err := c.conn.Modify(modReq); err != nil {
		return fmt.Errorf("failed to remove from group '%s': %w", groupName, err)
	}

	return nil
}

func (c *Client) ChangePassword(uid, newPassword string) error {
	searchReq := ldap.NewSearchRequest(
		c.cfg.PeopleOU,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(uid)),
		[]string{"dn"},
		nil,
	)

	result, err := c.conn.Search(searchReq)
	if err != nil {
		return fmt.Errorf("failed to search for user '%s': %w", uid, err)
	}

	if len(result.Entries) == 0 {
		return fmt.Errorf("user '%s' not found", uid)
	}

	userDN := result.Entries[0].DN

	modReq := ldap.NewModifyRequest(userDN, nil)
	modReq.Replace("userPassword", []string{newPassword})

	if err := c.conn.Modify(modReq); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

func (c *Client) ChangeEmail(uid, newEmail string) error {
	searchReq := ldap.NewSearchRequest(
		c.cfg.PeopleOU,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(uid)),
		[]string{"dn"},
		nil,
	)

	result, err := c.conn.Search(searchReq)
	if err != nil {
		return fmt.Errorf("failed to search for user '%s': %w", uid, err)
	}

	if len(result.Entries) == 0 {
		return fmt.Errorf("user '%s' not found", uid)
	}

	userDN := result.Entries[0].DN

	modReq := ldap.NewModifyRequest(userDN, nil)
	modReq.Replace("mail", []string{newEmail})

	if err := c.conn.Modify(modReq); err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	return nil
}

func (c *Client) GetNextGIDNumber() (int, error) {
	return c.nextIDNumber(c.cfg.GroupOU, "(objectClass=posixGroup)", "gidNumber", c.cfg.MinGIDNumber)
}

func (c *Client) CreateGroup(name string, gid int) error {
	groupDN := fmt.Sprintf("cn=%s,%s", name, c.cfg.GroupOU)

	addReq := ldap.NewAddRequest(groupDN, nil)
	addReq.Attribute("objectClass", []string{"posixGroup"})
	addReq.Attribute("cn", []string{name})
	addReq.Attribute("gidNumber", []string{strconv.Itoa(gid)})

	if err := c.conn.Add(addReq); err != nil {
		return fmt.Errorf("failed to create group '%s': %w", name, err)
	}

	return nil
}

func (c *Client) RemoveGroup(name string) error {
	groupDN := fmt.Sprintf("cn=%s,%s", name, c.cfg.GroupOU)

	delReq := ldap.NewDelRequest(groupDN, nil)

	if err := c.conn.Del(delReq); err != nil {
		return fmt.Errorf("failed to remove group '%s': %w", name, err)
	}

	return nil
}
