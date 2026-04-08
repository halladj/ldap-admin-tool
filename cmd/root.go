package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "ldap-admin-tool",
	Short:        "LDAP administration tool for misc-lab.org",
	Long:         "Manage LDAP users and groups. Create accounts, modify user details, and manage group memberships.",
	SilenceUsage: true, // don't print usage on LDAP/config errors
}

func init() {
	// Subcommands are added by their init() functions
}

func Execute() error {
	return rootCmd.Execute()
}
