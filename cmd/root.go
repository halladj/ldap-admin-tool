package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/misc-lab/ldap-user-tool/internal/config"
	ldapclient "github.com/misc-lab/ldap-user-tool/internal/ldap"
	"github.com/misc-lab/ldap-user-tool/internal/mail"
	"github.com/misc-lab/ldap-user-tool/internal/password"
	"github.com/misc-lab/ldap-user-tool/internal/pdf"
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

var rootCmd = &cobra.Command{
	Use:   "ldap-user-tool",
	Short: "LDAP user management tool for misc-lab.org",
	Long:  "Create LDAP user accounts, assign groups, generate credential PDFs, and send welcome emails.",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new LDAP user account",
	Example: `  ldap-user-tool create --first "Hamza" --last "Halladj" --email hamza@example.com
  ldap-user-tool create --first "Hamza" --last "Halladj" --email hamza@example.com --groups printing-a,admins-cups
  ldap-user-tool create --first "Hamza" --last "Halladj" --uid hhalladj --email hamza@example.com --password "MyP@ss123!" --no-email`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&firstName, "first", "", "First name (required)")
	createCmd.Flags().StringVar(&lastName, "last", "", "Last name (required)")
	createCmd.Flags().StringVar(&uid, "uid", "", "Username / uid (default: first letter of first name + last name)")
	createCmd.Flags().StringVar(&email, "email", "", "Email address (required)")
	createCmd.Flags().StringVar(&pass, "password", "", "Password (auto-generated if not provided)")
	createCmd.Flags().StringVar(&groups, "groups", "", "Comma-separated group names (e.g. printing-a,admins-cups)")
	createCmd.Flags().IntVar(&gid, "gid", 0, "Primary group ID (default: from config)")
	createCmd.Flags().BoolVar(&noEmail, "no-email", false, "Skip sending welcome email")
	createCmd.Flags().BoolVar(&noPDF, "no-pdf", false, "Skip PDF generation")

	createCmd.MarkFlagRequired("first")
	createCmd.MarkFlagRequired("last")
	createCmd.MarkFlagRequired("email")

	rootCmd.AddCommand(createCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Use config default GID if not set via flag
	if gid == 0 {
		gid = cfg.DefaultGID
	}

	// Load admin password
	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	// Generate uid if not provided: first letter of first name + last name
	if uid == "" {
		uid = strings.ToLower(string(firstName[0]) + lastName)
	}

	// Generate password if not provided
	userPass := pass
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

	// Get next UID number
	uidNumber, err := client.GetNextUIDNumber()
	if err != nil {
		return err
	}

	// Create user
	user := ldapclient.User{
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

	// Generate PDF
	var pdfPath string
	if !noPDF {
		pdfPath, err = pdf.Generate(cfg, pdf.UserInfo{
			FirstName: firstName,
			LastName:  lastName,
			UID:       uid,
			Email:     email,
			Password:  userPass,
			Groups:    groupList,
		})
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
	fmt.Printf("\n%s\n", strings.Repeat("=", 45))
	fmt.Printf("  Account created successfully!\n")
	fmt.Printf("  Username : %s\n", uid)
	fmt.Printf("  Password : %s\n", userPass)
	fmt.Printf("  Email    : %s\n", email)
	if len(groupList) > 0 {
		fmt.Printf("  Groups   : %s\n", strings.Join(groupList, ", "))
	}
	fmt.Printf("%s\n", strings.Repeat("=", 45))

	return nil
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
