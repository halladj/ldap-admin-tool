# ldap-admin-tool

A CLI tool for administering LDAP users and groups on your-domain.org. It automates user account creation, group management, credential PDF generation, and welcome email delivery.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [User Management](#user-management)
  - [Group Management](#group-management)
  - [Examples](#examples)
- [Architecture](#architecture)
- [Building from Source](#building-from-source)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- OpenLDAP server with TLS enabled
- Postfix (or compatible MTA) configured on the host for outbound email
- Admin credentials stored on the host (see [Configuration](#configuration))

## Installation

Copy the pre-built binary to the LDAP server:

```bash
scp ldap-admin-tool user@your-ldap-server:/tmp/
ssh user@your-ldap-server "sudo mv /tmp/ldap-admin-tool /usr/local/bin/ && sudo chmod +x /usr/local/bin/ldap-admin-tool"
```

Verify the installation:

```bash
ldap-admin-tool --help
```

## Configuration

The tool reads its LDAP admin password from a file on the host. Create it with restricted permissions:

```bash
echo -n "YOUR_ADMIN_PASSWORD" | sudo tee /etc/ldap/admin_pass
sudo chmod 600 /etc/ldap/admin_pass
```

### Default Settings

| Setting | Value |
|---------|-------|
| LDAP Server | `ldaps://ldap01.your-domain.org` |
| Base DN | `dc=your-domain,dc=org` |
| People OU | `ou=People,dc=your-domain,dc=org` |
| Group OU | `ou=group,dc=your-domain,dc=org` |
| Default GID | `10008` (jupyterhub-users-ldap) |
| Default Shell | `/bin/bash` |
| Sender Email | `no-replay@your-domain.org` |
| Admin Pass File | `/etc/ldap/admin_pass` |

To change defaults, use a config file at `/etc/ldap-admin-tool/config.yaml`, `~/.ldap-admin-tool/config.yaml`, or `./config.yaml` (see `config.yaml.example`).

Alternatively, set environment variables with the prefix `LDAP_ADMIN_TOOL_` (e.g., `LDAP_ADMIN_TOOL_LDAP_SERVER`).

## Usage

### User Management

#### Create a User

```bash
ldap-admin-tool user create --first <FIRST> --last <LAST> --email <EMAIL> [options]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--first` | Yes | - | First name |
| `--last` | Yes | - | Last name |
| `--email` | Yes | - | Email address |
| `--uid` | No | Auto | Username (first letter + last name, e.g. `ftotti`) |
| `--password` | No | Auto | Password (12-char strong if omitted) |
| `--groups` | No | - | Comma-separated group list (e.g. `printing-a,admins-cups`) |
| `--gid` | No | `10008` | Primary group ID |
| `--no-email` | No | `false` | Skip welcome email |
| `--no-pdf` | No | `false` | Skip PDF generation |

**Example:**

```bash
ldap-admin-tool user create \
  --first "Francesco" \
  --last "Totti" \
  --email jdoe@example.com \
  --groups printing-a,admins-cups
```

#### Modify a User

```bash
ldap-admin-tool user modify password --uid <UID> [password]
ldap-admin-tool user modify email --uid <UID> <EMAIL>
ldap-admin-tool user modify add-group --uid <UID> <GROUP> [group ...]
ldap-admin-tool user modify remove-group --uid <UID> <GROUP> [group ...]
```

**Examples:**

```bash
# Change password (auto-generated if omitted)
ldap-admin-tool user modify password --uid ftotti "NewP@ss1!"
ldap-admin-tool user modify password --uid ftotti

# Change email
ldap-admin-tool user modify email --uid ftotti new@your-domain.org

# Add to groups
ldap-admin-tool user modify add-group --uid ftotti printing-b admins-cups

# Remove from groups
ldap-admin-tool user modify remove-group --uid ftotti printing-a
```

### Group Management

#### Create a Group

```bash
ldap-admin-tool groups create <NAME> [--gid <GID>]
```

If `--gid` is omitted, the next available GID is auto-selected.

**Example:**

```bash
ldap-admin-tool groups create printing-c
ldap-admin-tool groups create printing-c --gid 10050
```

#### Remove a Group

```bash
ldap-admin-tool groups remove <NAME>
```

**Example:**

```bash
ldap-admin-tool groups remove printing-c
```

#### Add Users to a Group

```bash
ldap-admin-tool groups add-users <GROUP> <UID> [uid ...]
```

**Example:**

```bash
ldap-admin-tool groups add-users printing-a ftotti jdoe smithj
```

#### Remove Users from a Group

```bash
ldap-admin-tool groups remove-users <GROUP> <UID> [uid ...]
```

**Example:**

```bash
ldap-admin-tool groups remove-users printing-a ftotti jdoe
```

### Examples

**Create a user with auto-generated credentials:**

```bash
ldap-admin-tool user create \
  --first "Hamza" \
  --last "Halladj" \
  --email jdoe@example.com
```

**Create a user and skip email:**

```bash
ldap-admin-tool user create \
  --first "Test" \
  --last "User" \
  --email test@example.com \
  --no-email
```

## Architecture

```
ldap-admin-tool/
├── main.go                        # Entry point
├── cmd/
│   ├── root.go                    # Root command
│   ├── user.go                    # User command + create subcommand
│   ├── modify.go                  # User modify subcommand
│   └── groups.go                  # Groups commands
├── internal/
│   ├── config/
│   │   └── config.go              # Viper-based configuration
│   ├── ldap/
│   │   └── client.go              # LDAP connection and operations
│   ├── mail/
│   │   └── sender.go              # Email with PDF attachment
│   ├── password/
│   │   └── generator.go           # Secure password generation
│   └── pdf/
│       └── generator.go           # Credential PDF generation
├── go.mod
└── go.sum
```

### Package Responsibilities

| Package | Responsibility |
|---------|-----------------|
| `cmd` | CLI interface (Cobra). Command parsing and orchestration. |
| `internal/config` | Viper-based configuration from files/environment. Admin password loading. |
| `internal/ldap` | LDAP client. TLS connection, user/group CRUD, UID/GID allocation. |
| `internal/mail` | MIME multipart email construction. Sendmail integration. |
| `internal/password` | Cryptographically secure password generation with complexity enforcement. |
| `internal/pdf` | Branded PDF credential document generation (gofpdf). |

### Dependencies

| Dependency | Purpose |
|------------|---------|
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [github.com/spf13/viper](https://github.com/spf13/viper) | Configuration management |
| [github.com/go-ldap/ldap/v3](https://github.com/go-ldap/ldap) | LDAP client library |
| [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) | PDF generation |

## Building from Source

**Requirements:** Go 1.24+

```bash
git clone https://github.com/halladj/ldap-admin-tool.git
cd ldap-admin-tool
go mod tidy
go build -o ldap-admin-tool .
```

### Cross-compile for Linux amd64:

```bash
GOOS=linux GOARCH=amd64 go build -o ldap-admin-tool .
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `cannot read /etc/ldap/admin_pass` | Missing or wrong permissions | Create file and set `chmod 600` |
| `failed to connect to LDAP` | Server unreachable or TLS issue | Verify LDAP server is accessible |
| `failed to bind as admin` | Wrong admin password | Check `/etc/ldap/admin_pass` content |
| `failed to create user` | Duplicate CN/UID or OU missing | Verify user doesn't exist; check LDAP tree |
| `group 'X' not found` | Group doesn't exist | List groups with `ldapsearch` |
| `sendmail failed` | Postfix not running | Run `sudo systemctl start postfix` |

## License

Internal tool for your-domain.org.
