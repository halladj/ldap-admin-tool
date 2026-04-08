package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/misc-lab/ldap-user-tool/internal/config"
	ldapclient "github.com/misc-lab/ldap-user-tool/internal/ldap"
	"github.com/misc-lab/ldap-user-tool/internal/password"
)

var (
	passwdUID  string
	passwdPass string
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change an LDAP user's password",
	Example: `  ldap-user-tool passwd --uid ftotti
  ldap-user-tool passwd --uid ftotti --password "NewP@ss123!"`,
	RunE: runPasswd,
}

func init() {
	passwdCmd.Flags().StringVar(&passwdUID, "uid", "", "Username / uid (required)")
	passwdCmd.Flags().StringVar(&passwdPass, "password", "", "New password (auto-generated if not provided)")

	passwdCmd.MarkFlagRequired("uid")

	rootCmd.AddCommand(passwdCmd)
}

func runPasswd(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Load admin password
	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	// Generate password if not provided
	userPass := passwdPass
	if userPass == "" {
		userPass, err = password.Generate(12)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	}

	// Connect to LDAP
	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	defer client.Close()

	// Change password
	if err := client.ChangePassword(passwdUID, userPass); err != nil {
		return err
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Password changed successfully!\n")
	fmt.Printf("  Username : %s\n", passwdUID)
	fmt.Printf("  Password : %s\n", userPass)
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
}
