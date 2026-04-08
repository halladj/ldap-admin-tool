package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ldap-admin-tool",
	Short: "LDAP administration tool for your-domain.org",
	Long:  "Manage LDAP users and groups. Create accounts, modify user details, and manage group memberships.",
}

func init() {
	// Subcommands are added by their init() functions
}

func Execute() error {
	return rootCmd.Execute()
}
