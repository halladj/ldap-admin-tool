package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/halladj/ldap-admin-tool/internal/config"
	ldapclient "github.com/halladj/ldap-admin-tool/internal/ldap"
)

var groupsGID int

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage LDAP groups",
	Long:  "Create, delete, and manage group memberships.",
}

var groupsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new LDAP group",
	Args:  cobra.ExactArgs(1),
	Example: `  ldap-admin-tool groups create printing-c
  ldap-admin-tool groups create printing-c --gid 10050`,
	RunE: runGroupsCreate,
}

var groupsRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an LDAP group",
	Args:  cobra.ExactArgs(1),
	Example: `  ldap-admin-tool groups remove printing-c`,
	RunE: runGroupsRemove,
}

var groupsAddUsersCmd = &cobra.Command{
	Use:   "add-users <group> <user> [user ...]",
	Short: "Add one or more users to a group",
	Args:  cobra.MinimumNArgs(2),
	Example: `  ldap-admin-tool groups add-users printing-a ftotti
  ldap-admin-tool groups add-users printing-a ftotti jdoe smithj`,
	RunE: runGroupsAddUsers,
}

var groupsRemoveUsersCmd = &cobra.Command{
	Use:   "remove-users <group> <user> [user ...]",
	Short: "Remove one or more users from a group",
	Args:  cobra.MinimumNArgs(2),
	Example: `  ldap-admin-tool groups remove-users printing-a ftotti
  ldap-admin-tool groups remove-users printing-a ftotti jdoe`,
	RunE: runGroupsRemoveUsers,
}

func init() {
	groupsCreateCmd.Flags().IntVar(&groupsGID, "gid", 0, "Group ID (auto-selected if not provided)")

	groupsCmd.AddCommand(groupsCreateCmd)
	groupsCmd.AddCommand(groupsRemoveCmd)
	groupsCmd.AddCommand(groupsAddUsersCmd)
	groupsCmd.AddCommand(groupsRemoveUsersCmd)

	rootCmd.AddCommand(groupsCmd)
}

func runGroupsCreate(cmd *cobra.Command, args []string) error {
	groupName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	// If GID not provided, get next available
	gid := groupsGID
	if gid == 0 {
		gid, err = client.GetNextGIDNumber()
		if err != nil {
			return err
		}
	}

	if err := client.CreateGroup(groupName, gid); err != nil {
		return err
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Group created successfully!\n")
	fmt.Printf("  Group name : %s\n", groupName)
	fmt.Printf("  GID        : %d\n", gid)
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runGroupsRemove(cmd *cobra.Command, args []string) error {
	groupName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.RemoveGroup(groupName); err != nil {
		return err
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Group removed successfully!\n")
	fmt.Printf("  Group name : %s\n", groupName)
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runGroupsAddUsers(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	users := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Adding users to group '%s'...\n", groupName)

	for _, uid := range users {
		if err := client.AddToGroup(uid, groupName); err != nil {
			fmt.Printf("  [!] %v\n", err)
		} else {
			fmt.Printf("  [+] Added '%s' to group\n", uid)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runGroupsRemoveUsers(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	users := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Removing users from group '%s'...\n", groupName)

	for _, uid := range users {
		if err := client.RemoveFromGroup(uid, groupName); err != nil {
			fmt.Printf("  [!] %v\n", err)
		} else {
			fmt.Printf("  [-] Removed '%s' from group\n", uid)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}
