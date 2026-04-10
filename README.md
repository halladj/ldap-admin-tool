# ldap-admin-tool

A CLI tool for administering LDAP users and groups. It automates user account creation, group management, credential PDF generation, and welcome email delivery.

## Prerequisites

- OpenLDAP server with TLS enabled
- Postfix (or compatible MTA) configured on the host for outbound email
- Admin credentials stored on the host (see Configuration)

## Installation

```bash
scp ldap-admin-tool user@your-ldap-server:/tmp/
ssh user@your-ldap-server "sudo mv /tmp/ldap-admin-tool /usr/local/bin/ && sudo chmod +x /usr/local/bin/ldap-admin-tool"
```

## Configuration

```bash
echo -n "YOUR_ADMIN_PASSWORD" | sudo tee /etc/ldap/admin_pass
sudo chmod 600 /etc/ldap/admin_pass
```

Config file locations (in order of precedence): `./config.yaml`, `~/.ldap-admin-tool/config.yaml`, `/etc/ldap-admin-tool/config.yaml`. See `config.yaml.example`.

Environment variables are supported with the prefix `LDAP_ADMIN_TOOL_` (e.g. `LDAP_ADMIN_TOOL_LDAP_SERVER`).

## Usage

### Command Hierarchy

```
ldap-admin-tool
├── user
│   ├── create              Create new LDAP user account
│   │   └── Flags: --first*, --last*, --email*
│   │           --uid, --password, --groups, --gid
│   │           --no-email, --no-pdf
│   ├── query               List all users, or show one with --uid
│   ├── delete              Delete a user (--uid required)
│   └── modify
│       ├── password [PWD]  Change/reset password
│       ├── email <EMAIL>   Update email address
│       ├── add-group       Add to one or more groups
│       └── remove-group    Remove from one or more groups
│
└── groups
    ├── create <NAME>       Create new group
    ├── remove <NAME>       Delete group
    ├── query [NAME]        List all groups, or show one by name
    ├── add-users           Add users to group
    └── remove-users        Remove users from group
```

### User Management

```bash
# Create
ldap-admin-tool user create --first "John" --last "Doe" --email jdoe@example.com --groups printing-a

# Query
ldap-admin-tool user query              # list all
ldap-admin-tool user query --uid jdoe  # show one

# Delete
ldap-admin-tool user delete --uid jdoe

# Modify
ldap-admin-tool user modify password --uid jdoe "NewP@ss1!"
ldap-admin-tool user modify password --uid jdoe        # auto-generate
ldap-admin-tool user modify email --uid jdoe new@example.com
ldap-admin-tool user modify add-group --uid jdoe printing-b admins-cups
ldap-admin-tool user modify remove-group --uid jdoe printing-a
```

### Group Management

```bash
ldap-admin-tool groups create printing-c
ldap-admin-tool groups create printing-c --gid 10050
ldap-admin-tool groups remove printing-c
ldap-admin-tool groups query                   # list all
ldap-admin-tool groups query printing-a        # show one
ldap-admin-tool groups add-users printing-a jdoe jsmith
ldap-admin-tool groups remove-users printing-a jdoe jsmith
```

### Shell Autocompletion

```bash
ldap-admin-tool completion bash | sudo tee /etc/bash_completion.d/ldap-admin-tool
ldap-admin-tool completion zsh > "${fpath[1]}/_ldap-admin-tool"
ldap-admin-tool completion fish | source
```

## Architecture

```
ldap-admin-tool/
├── main.go
├── cmd/
│   ├── root.go            # Root command
│   ├── ldap_helper.go     # LDAP connection bootstrap
│   ├── output.go          # Output formatting helpers
│   ├── user.go            # User create command
│   ├── modify.go          # User modify subcommands
│   ├── query.go           # User query/delete commands
│   └── groups.go          # Groups commands
├── internal/
│   ├── config/config.go   # Viper-based configuration
│   ├── ldap/client.go     # LDAP connection and operations
│   ├── mail/sender.go     # Email with PDF attachment
│   ├── password/          # Secure password generation
│   ├── pdf/
│   │   ├── generator.go   # PDF generation
│   │   └── assets/logo.png
│   └── types/user.go      # Shared types
├── go.mod
└── go.sum
```

### Dependencies

| Dependency | Purpose |
|------------|---------|
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [github.com/spf13/viper](https://github.com/spf13/viper) | Configuration management |
| [github.com/go-ldap/ldap/v3](https://github.com/go-ldap/ldap) | LDAP client library |
| [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) | PDF generation |

## Building from Source

**Requirements:** Go 1.21+

```bash
git clone https://github.com/halladj/ldap-admin-tool.git
cd ldap-admin-tool
go mod tidy
go build -o ldap-admin-tool .
```

Cross-compile for Linux amd64:

```bash
GOOS=linux GOARCH=amd64 go build -o ldap-admin-tool .
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `cannot read /etc/ldap/admin_pass` | Missing or wrong permissions | `chmod 600 /etc/ldap/admin_pass` |
| `failed to connect to LDAP` | Server unreachable or TLS issue | Verify LDAP server is accessible |
| `failed to bind as admin` | Wrong admin password | Check `/etc/ldap/admin_pass` |
| `failed to create user` | Duplicate CN/UID or missing OU | Verify user doesn't exist |
| `group 'X' not found` | Group doesn't exist | Use `groups query` to list groups |
| `sendmail failed` | Postfix not running | `sudo systemctl start postfix` |

## License

MIT
