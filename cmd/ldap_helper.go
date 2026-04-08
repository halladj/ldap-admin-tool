package cmd

import (
	"github.com/misc-lab/ldap-admin-tool/internal/config"
	"github.com/misc-lab/ldap-admin-tool/internal/ldap"
)

// withLDAPClient establishes an LDAP connection and passes it to a callback function.
// The connection is automatically closed when the callback returns.
func withLDAPClient(fn func(*config.Config, *ldap.Client) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	client, err := ldap.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	return fn(cfg, client)
}
