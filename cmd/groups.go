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

var groupsQueryCmd = &cobra.Command{
	Use:   "query [name]",
	Short: "Query an LDAP group, or list all groups if no name given",
	Args:  cobra.MaximumNArgs(1),
	Example: `  ldap-admin-tool groups query             # list all groups
  ldap-admin-tool groups query printing-a  # show one group`,
	RunE: runGroupsQuery,
}

func init() {
	groupsCreateCmd.Flags().IntVar(&groupsGID, "gid", 0, "Group ID (auto-selected if not provided)")

	groupsCmd.AddCommand(groupsCreateCmd)
	groupsCmd.AddCommand(groupsRemoveCmd)
	groupsCmd.AddCommand(groupsAddUsersCmd)
	groupsCmd.AddCommand(groupsRemoveUsersCmd)
	groupsCmd.AddCommand(groupsQueryCmd)

	rootCmd.AddCommand(groupsCmd)
}

func runGroupsCreate(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		// If GID not provided, get next available
		gid := groupsGID
		if gid == 0 {
			var err error
			gid, err = client.GetNextGIDNumber()
			if err != nil {
				return err
			}
		}

		if err := client.CreateGroup(groupName, gid); err != nil {
			return err
		}

		printBanner("Group created successfully!",
			"Group name", groupName,
			"GID", fmt.Sprintf("%d", gid))

		return nil
	})
}

func runGroupsRemove(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		if err := client.RemoveGroup(groupName); err != nil {
			return err
		}

		printBanner("Group removed successfully!",
			"Group name", groupName)

		return nil
	})
}

func runGroupsAddUsers(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	users := args[1:]
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		printProgress(fmt.Sprintf("Adding users to group '%s'...", groupName))

		for _, uid := range users {
			if err := client.AddToGroup(uid, groupName); err != nil {
				fmt.Printf("  [!] %v\n", err)
			} else {
				fmt.Printf("  [+] Added '%s' to group\n", uid)
			}
		}

		printDone()

		return nil
	})
}

func runGroupsRemoveUsers(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	users := args[1:]
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		printProgress(fmt.Sprintf("Removing users from group '%s'...", groupName))

		for _, uid := range users {
			if err := client.RemoveFromGroup(uid, groupName); err != nil {
				fmt.Printf("  [!] %v\n", err)
			} else {
				fmt.Printf("  [-] Removed '%s' from group\n", uid)
			}
		}

		printDone()

		return nil
	})
}

func runGroupsQuery(cmd *cobra.Command, args []string) error {
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		if len(args) == 1 {
			g, err := client.QueryGroup(args[0])
			if err != nil {
				return err
			}
			printGroupDetails(g)
			return nil
		}

		groups, err := client.ListGroups()
		if err != nil {
			return err
		}

		fmt.Printf("\n%-30s %-10s %-s\n", "Group", "GID", "Members")
		fmt.Println(strings.Repeat("-", 62))
		for _, g := range groups {
			members := strings.Join(g.Members, ", ")
			if members == "" {
				members = "-"
			}
			fmt.Printf("%-30s %-10d %s\n", g.Name, g.GID, members)
		}
		fmt.Printf("\nTotal: %d group(s)\n", len(groups))

		return nil
	})
}
