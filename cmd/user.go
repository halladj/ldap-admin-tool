package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/misc-lab/ldap-admin-tool/internal/config"
	ldapclient "github.com/misc-lab/ldap-admin-tool/internal/ldap"
	"github.com/misc-lab/ldap-admin-tool/internal/mail"
	"github.com/misc-lab/ldap-admin-tool/internal/password"
	"github.com/misc-lab/ldap-admin-tool/internal/pdf"
	"github.com/misc-lab/ldap-admin-tool/internal/types"
)

var (
	firstName string
	lastName  string
	uid       string
	email     string
	pass      string
	groups    string
	gid       int
	noEmail   bool
	noPDF     bool
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage LDAP user accounts",
	Long:  "Create and modify LDAP user accounts.",
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new LDAP user account",
	Example: `  ldap-admin-tool user create --first "Hamza" --last "Halladj" --email hamza@example.com
  ldap-admin-tool user create --first "Hamza" --last "Halladj" --email hamza@example.com --groups printing-a,admins-cups
  ldap-admin-tool user create --first "Hamza" --last "Halladj" --uid hhalladj --email hamza@example.com --password "MyP@ss123!" --no-email`,
	RunE: runUserCreate,
}

func init() {
	userCreateCmd.Flags().StringVar(&firstName, "first", "", "First name (required)")
	userCreateCmd.Flags().StringVar(&lastName, "last", "", "Last name (required)")
	userCreateCmd.Flags().StringVar(&uid, "uid", "", "Username / uid (default: first letter of first name + last name)")
	userCreateCmd.Flags().StringVar(&email, "email", "", "Email address (required)")
	userCreateCmd.Flags().StringVar(&pass, "password", "", "Password (auto-generated if not provided)")
	userCreateCmd.Flags().StringVar(&groups, "groups", "", "Comma-separated group names (e.g. printing-a,admins-cups)")
	userCreateCmd.Flags().IntVar(&gid, "gid", 0, "Primary group ID (default: from config)")
	userCreateCmd.Flags().BoolVar(&noEmail, "no-email", false, "Skip sending welcome email")
	userCreateCmd.Flags().BoolVar(&noPDF, "no-pdf", false, "Skip PDF generation")

	userCreateCmd.MarkFlagRequired("first")
	userCreateCmd.MarkFlagRequired("last")
	userCreateCmd.MarkFlagRequired("email")

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(modifyCmd)

	rootCmd.AddCommand(userCmd)
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	return withLDAPClient(func(cfg *config.Config, client *ldapclient.Client) error {
		// Use config default GID if not set via flag
		if gid == 0 {
			gid = cfg.DefaultGID
		}

		// Generate uid if not provided: first letter of first name + last name
		if uid == "" {
			uid = strings.ToLower(string(firstName[0]) + lastName)
		}

		// Generate password if not provided
		userPass := pass
		if userPass == "" {
			var err error
			userPass, err = password.Generate(12)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}
		}

		// Get next UID number
		uidNumber, err := client.GetNextUIDNumber()
		if err != nil {
			return err
		}

		// Create user
		user := types.User{
			FirstName: firstName,
			LastName:  lastName,
			UID:       uid,
			Email:     email,
			Password:  userPass,
			GID:       gid,
		}

		_, err = client.CreateUser(user, uidNumber)
		if err != nil {
			return err
		}
		fmt.Printf("[+] User '%s' created (uidNumber: %d)\n", uid, uidNumber)

		// Add to groups
		groupList := parseGroups(groups)
		for _, g := range groupList {
			if err := client.AddToGroup(uid, g); err != nil {
				fmt.Fprintf(os.Stderr, "[!] %v\n", err)
			} else {
				fmt.Printf("[+] Added '%s' to group '%s'\n", uid, g)
			}
		}
		user.Groups = groupList

		// Generate PDF
		var pdfPath string
		if !noPDF {
			pdfPath, err = pdf.Generate(cfg, user)
			if err != nil {
				return fmt.Errorf("failed to generate PDF: %w", err)
			}
			fmt.Printf("[+] PDF generated: %s\n", pdfPath)
			defer os.Remove(pdfPath)
		}

		// Send email
		if !noEmail && pdfPath != "" {
			if err := mail.SendWelcome(cfg.SenderEmail, email, firstName, lastName, uid, pdfPath); err != nil {
				return fmt.Errorf("failed to send email: %w", err)
			}
			fmt.Printf("[+] Welcome email sent to %s\n", email)
		}

		// Summary
		rows := []string{
			"Username", uid,
			"Password", userPass,
			"Email", email,
		}
		if len(groupList) > 0 {
			rows = append(rows, "Groups", strings.Join(groupList, ", "))
		}
		printBanner("Account created successfully!", rows...)

		return nil
	})
}

func parseGroups(g string) []string {
	if g == "" {
		return nil
	}
	parts := strings.Split(g, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
