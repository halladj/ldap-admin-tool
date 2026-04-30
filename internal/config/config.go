package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LDAPServer    string `mapstructure:"ldap_server"`
	BaseDN        string `mapstructure:"base_dn"`
	PeopleOU      string `mapstructure:"people_ou"`
	GroupOU       string `mapstructure:"group_ou"`
	AdminDN       string `mapstructure:"admin_dn"`
	AdminPassFile string `mapstructure:"admin_pass_file"`
	SenderEmail   string `mapstructure:"sender_email"`
	DefaultShell  string `mapstructure:"default_shell"`
	DefaultGID    int    `mapstructure:"default_gid"`
	MinUIDNumber  int    `mapstructure:"min_uid_number"`
	MinGIDNumber  int    `mapstructure:"min_gid_number"`
	APIKey        string `mapstructure:"api_key"`
}

// Build-time injectable defaults (override with -ldflags "-X github.com/halladj/ldap-admin-tool/internal/config.DefaultLDAPServer=...")
var (
	DefaultLDAPServer   = "ldaps://ldap.your-domain.org"
	DefaultBaseDN       = "dc=your-domain,dc=org"
	DefaultPeopleOU     = "ou=People,dc=your-domain,dc=org"
	DefaultGroupOU      = "ou=group,dc=your-domain,dc=org"
	DefaultAdminDN      = "cn=admin,dc=your-domain,dc=org"
	DefaultAdminPassFile = "/etc/ldap/admin_pass"
	DefaultSenderEmail  = "noreply@your-domain.org"
	DefaultShell        = "/bin/bash"
)

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/ldap-admin-tool")
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".ldap-admin-tool"))
	}
	viper.AddConfigPath(".")

	// Defaults (overridable at build time via -ldflags)
	viper.SetDefault("ldap_server", DefaultLDAPServer)
	viper.SetDefault("base_dn", DefaultBaseDN)
	viper.SetDefault("people_ou", DefaultPeopleOU)
	viper.SetDefault("group_ou", DefaultGroupOU)
	viper.SetDefault("admin_dn", DefaultAdminDN)
	viper.SetDefault("admin_pass_file", DefaultAdminPassFile)
	viper.SetDefault("sender_email", DefaultSenderEmail)
	viper.SetDefault("default_shell", DefaultShell)
	viper.SetDefault("default_gid", 10008)
	viper.SetDefault("min_uid_number", 10000)
	viper.SetDefault("min_gid_number", 10000)

	// Environment variables: LDAP_ADMIN_TOOL_LDAP_SERVER, etc.
	viper.SetEnvPrefix("LDAP_ADMIN_TOOL")
	viper.AutomaticEnv()

	// Read config file (optional — defaults work without one)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) LoadAdminPassword() (string, error) {
	data, err := os.ReadFile(c.AdminPassFile)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", c.AdminPassFile, err)
	}
	return strings.TrimSpace(string(data)), nil
}
