package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/misc-lab/ldap-admin-tool/internal/config"
	ldapclient "github.com/misc-lab/ldap-admin-tool/internal/ldap"
	"github.com/misc-lab/ldap-admin-tool/internal/password"
)

var modifyUID string

var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify an LDAP user account",
	Long:  "Modify user password, email, and group memberships.",
}

var modifyPasswordCmd = &cobra.Command{
	Use:   "password [password]",
	Short: "Change a user's password",
	Long:  "Change a user's password. If no password is provided, a new one will be auto-generated.",
	Args:  cobra.MaximumNArgs(1),
	Example: `  ldap-admin-tool user modify password --uid ftotti "NewP@ss1!"
  ldap-admin-tool user modify password --uid ftotti  # auto-generates`,
	RunE: runModifyPassword,
}

var modifyEmailCmd = &cobra.Command{
	Use:   "email <email>",
	Short: "Change a user's email",
	Args:  cobra.ExactArgs(1),
	Example: `  ldap-admin-tool user modify email --uid ftotti new@misc-lab.org`,
	RunE: runModifyEmail,
}

var modifyAddGroupCmd = &cobra.Command{
	Use:   "add-group <group> [group ...]",
	Short: "Add a user to one or more groups",
	Args:  cobra.MinimumNArgs(1),
	Example: `  ldap-admin-tool user modify add-group --uid ftotti printing-b
  ldap-admin-tool user modify add-group --uid ftotti printing-b admins-cups`,
	RunE: runModifyAddGroup,
}

var modifyRemoveGroupCmd = &cobra.Command{
	Use:   "remove-group <group> [group ...]",
	Short: "Remove a user from one or more groups",
	Args:  cobra.MinimumNArgs(1),
	Example: `  ldap-admin-tool user modify remove-group --uid ftotti printing-a
  ldap-admin-tool user modify remove-group --uid ftotti printing-a printing-b`,
	RunE: runModifyRemoveGroup,
}

func init() {
	modifyCmd.PersistentFlags().StringVar(&modifyUID, "uid", "", "Username / uid (required)")
	modifyCmd.MarkPersistentFlagRequired("uid")

	modifyCmd.AddCommand(modifyPasswordCmd)
	modifyCmd.AddCommand(modifyEmailCmd)
	modifyCmd.AddCommand(modifyAddGroupCmd)
	modifyCmd.AddCommand(modifyRemoveGroupCmd)
}

func runModifyPassword(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	userPass := ""
	if len(args) > 0 {
		userPass = args[0]
	} else {
		userPass, err = password.Generate(12)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	}

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.ChangePassword(modifyUID, userPass); err != nil {
		return err
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Password changed successfully!\n")
	fmt.Printf("  Username : %s\n", modifyUID)
	fmt.Printf("  Password : %s\n", userPass)
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runModifyEmail(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	newEmail := args[0]

	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.ChangeEmail(modifyUID, newEmail); err != nil {
		return err
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Email changed successfully!\n")
	fmt.Printf("  Username : %s\n", modifyUID)
	fmt.Printf("  Email    : %s\n", newEmail)
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runModifyAddGroup(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("  Adding user to groups...\n")

	for _, group := range args {
		if err := client.AddToGroup(modifyUID, group); err != nil {
			fmt.Printf("  [!] %v\n", err)
		} else {
			fmt.Printf("  [+] Added to group '%s'\n", group)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}

func runModifyRemoveGroup(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("  Removing user from groups...\n")

	for _, group := range args {
		if err := client.RemoveFromGroup(modifyUID, group); err != nil {
			fmt.Printf("  [!] %v\n", err)
		} else {
			fmt.Printf("  [-] Removed from group '%s'\n", group)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}
