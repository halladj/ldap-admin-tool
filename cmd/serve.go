package cmd

import (
	"github.com/halladj/ldap-admin-tool/api"
	"github.com/halladj/ldap-admin-tool/internal/config"
	ldapclient "github.com/halladj/ldap-admin-tool/internal/ldap"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server with Swagger UI",
	Long:  "Starts an HTTP server exposing all LDAP operations as a REST API. Swagger UI available at /swagger/index.html.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

func runServe(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	adminPass, err := cfg.LoadAdminPassword()
	if err != nil {
		return err
	}

	// Validate LDAP connectivity before starting the server
	client, err := ldapclient.NewClient(cfg, adminPass)
	if err != nil {
		return err
	}
	client.Close()

	return api.NewServer(cfg, adminPass).Start(servePort)
}
