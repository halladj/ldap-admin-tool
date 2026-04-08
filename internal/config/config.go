package config

import (
	"fmt"
	"os"
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
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/ldap-user-tool")
	viper.AddConfigPath("$HOME/.ldap-user-tool")
	viper.AddConfigPath(".")

	// Defaults
	viper.SetDefault("ldap_server", "ldaps://ldap01.your-domain.org")
	viper.SetDefault("base_dn", "dc=your-domain,dc=org")
	viper.SetDefault("people_ou", "ou=People,dc=your-domain,dc=org")
	viper.SetDefault("group_ou", "ou=group,dc=your-domain,dc=org")
	viper.SetDefault("admin_dn", "cn=admin,dc=your-domain,dc=org")
	viper.SetDefault("admin_pass_file", "/etc/ldap/admin_pass")
	viper.SetDefault("sender_email", "no-replay@your-domain.org")
	viper.SetDefault("default_shell", "/bin/bash")
	viper.SetDefault("default_gid", 10008)

	// Environment variables: LDAP_USER_TOOL_LDAP_SERVER, etc.
	viper.SetEnvPrefix("LDAP_USER_TOOL")
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
