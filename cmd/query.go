package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/misc-lab/ldap-admin-tool/internal/config"
	ldapclient "github.com/misc-lab/ldap-admin-tool/internal/ldap"
	"github.com/misc-lab/ldap-admin-tool/internal/types"
)

var queryUID string

var userQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query LDAP user accounts",
	Example: `  ldap-admin-tool user query            # list all users
  ldap-admin-tool user query --uid ftotti  # show one user`,
	RunE: runUserQuery,
}

var userDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an LDAP user account",
	Example: `  ldap-admin-tool user delete --uid ftotti`,
	RunE:    runUserDelete,
}

func init() {
	userQueryCmd.Flags().StringVar(&queryUID, "uid", "", "Username / uid (omit to list all users)")

	userDeleteCmd.Flags().StringVar(&queryUID, "uid", "", "Username / uid (required)")
	userDeleteCmd.MarkFlagRequired("uid")

	userCmd.AddCommand(userQueryCmd)
	userCmd.AddCommand(userDeleteCmd)
}

func runUserQuery(cmd *cobra.Command, args []string) error {
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		if queryUID != "" {
			return printUserDetails(client, queryUID)
		}

		users, err := client.ListUsers()
		if err != nil {
			return err
		}

		fmt.Printf("\n%-20s %-10s %-30s\n", "UID", "UID Number", "Email")
		fmt.Println(strings.Repeat("-", 62))
		for _, u := range users {
			fmt.Printf("%-20s %-10d %-30s\n", u.UID, u.UIDNumber, u.Email)
		}
		fmt.Printf("\nTotal: %d user(s)\n", len(users))

		return nil
	})
}

func printUserDetails(client *ldapclient.Client, uid string) error {
	u, err := client.QueryUser(uid)
	if err != nil {
		return err
	}

	groups := "none"
	if len(u.Groups) > 0 {
		groups = strings.Join(u.Groups, ", ")
	}

	printBanner("User: "+u.UID,
		"DN", u.DN,
		"Full Name", u.FirstName+" "+u.LastName,
		"Email", u.Email,
		"UID Number", fmt.Sprintf("%d", u.UIDNumber),
		"GID Number", fmt.Sprintf("%d", u.GIDNumber),
		"Home Dir", u.HomeDir,
		"Shell", u.Shell,
		"Groups", groups,
	)

	return nil
}

func printGroupDetails(g *types.GroupDetails) {
	members := "none"
	if len(g.Members) > 0 {
		members = strings.Join(g.Members, ", ")
	}

	printBanner("Group: "+g.Name,
		"DN", g.DN,
		"GID", fmt.Sprintf("%d", g.GID),
		"Members", members,
	)
}

func runUserDelete(cmd *cobra.Command, args []string) error {
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		if err := client.DeleteUser(queryUID); err != nil {
			return err
		}

		printBanner("User deleted successfully!",
			"Username", queryUID)

		return nil
	})
}

