package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
)

var generateAPIKeyCmd = &cobra.Command{
	Use:   "generate-apikey",
	Short: "Generate a secure random API key",
	Long:  "Generates a cryptographically secure 32-byte (64-char hex) API key. Add it to config.yaml as api_key or export as LDAP_ADMIN_TOOL_API_KEY.",
	RunE: func(_ *cobra.Command, _ []string) error {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
		key := hex.EncodeToString(b)
		fmt.Println(key)
		fmt.Printf("\nAdd to config.yaml:\n  api_key: \"%s\"\n\nOr export:\n  export LDAP_ADMIN_TOOL_API_KEY=%s\n", key, key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateAPIKeyCmd)
}
