# ldap-user-tool

A CLI tool for managing LDAP user accounts on misc-lab.org. It automates account creation, group assignment, credential PDF generation, and welcome email delivery.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Create a User](#create-a-user)
  - [Options](#options)
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
scp ldap-user-tool misc@10.0.10.8:/tmp/
ssh misc@10.0.10.8 "sudo mv /tmp/ldap-user-tool /usr/local/bin/ && sudo chmod +x /usr/local/bin/ldap-user-tool"
```

Verify the installation:

```bash
ldap-user-tool --help
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
| LDAP Server | `ldaps://ldap01.misc-lab.org` |
| Base DN | `dc=misc-lab,dc=org` |
| People OU | `ou=People,dc=misc-lab,dc=org` |
| Group OU | `ou=group,dc=misc-lab,dc=org` |
| Default GID | `10008` (jupyterhub-users-ldap) |
| Default Shell | `/bin/bash` |
| Sender Email | `no-replay@misc-lab.org` |
| Admin Pass File | `/etc/ldap/admin_pass` |

To change defaults, edit `internal/config/config.go` and rebuild.

## Usage

### Create a User

```bash
ldap-user-tool create --first <FIRST_NAME> --last <LAST_NAME> --email <EMAIL> [options]
```

### Options

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--first` | Yes | - | First name |
| `--last` | Yes | - | Last name |
| `--email` | Yes | - | Email address |
| `--uid` | No | Auto-generated | Username. Defaults to first letter of first name + last name (e.g. `ftotti`) |
| `--password` | No | Auto-generated | Password. If omitted, a 12-character strong password is generated |
| `--groups` | No | - | Comma-separated list of LDAP groups to add the user to |
| `--gid` | No | `10008` | Primary group ID |
| `--no-email` | No | `false` | Skip sending the welcome email |
| `--no-pdf` | No | `false` | Skip PDF credential generation |

### Examples

**Create a user with auto-generated uid and password:**

```bash
ldap-user-tool create \
  --first "Hamza" \
  --last "Halladj" \
  --email hamza@example.com
```

Output:
```
[+] User 'hhalladj' created (uidNumber: 10036)
[+] PDF generated: /tmp/ldap_hhalladj_1234567890.pdf
[+] Welcome email sent to hamza@example.com

=============================================
  Account created successfully!
  Username : hhalladj
  Password : aB3$kLm9!xYz
  Email    : hamza@example.com
=============================================
```

**Create a user with a specific password and groups:**

```bash
ldap-user-tool create \
  --first "Francesco" \
  --last "Totti" \
  --email ftotti@example.com \
  --password "MyP@ssw0rd!" \
  --groups printing-a,admins-cups
```

**Create a user with a custom uid:**

```bash
ldap-user-tool create \
  --first "Ahmed" \
  --last "Chaouche" \
  --uid achaouche \
  --email ahmed@example.com
```

**Create a user without sending an email:**

```bash
ldap-user-tool create \
  --first "Test" \
  --last "User" \
  --email test@example.com \
  --no-email
```

## Architecture

```
ldap-user-tool/
├── main.go                        # Entry point
├── cmd/
│   └── root.go                    # CLI commands (cobra)
├── internal/
│   ├── config/
│   │   └── config.go              # Server settings and admin password
│   ├── ldap/
│   │   └── client.go              # LDAP connection, user creation, group management
│   ├── mail/
│   │   └── sender.go              # Email with PDF attachment via sendmail
│   ├── password/
│   │   └── generator.go           # Cryptographically secure password generation
│   └── pdf/
│       └── generator.go           # Credential PDF generation
├── go.mod
└── go.sum
```

### Package Responsibilities

| Package | Description |
|---------|-------------|
| `cmd` | CLI interface using [Cobra](https://github.com/spf13/cobra). Parses flags, orchestrates the creation workflow. |
| `internal/config` | Centralized configuration constants and admin password loading. |
| `internal/ldap` | LDAP client. Handles TLS connection, user creation, UID allocation, and group membership. |
| `internal/mail` | Constructs MIME multipart emails with PDF attachments. Sends via local `sendmail`. |
| `internal/password` | Generates cryptographically secure passwords using `crypto/rand`. Enforces complexity (upper, lower, digit, special). |
| `internal/pdf` | Generates branded PDF documents containing user credentials using [gofpdf](https://github.com/jung-kurt/gofpdf). |

### Dependencies

| Dependency | Purpose |
|------------|---------|
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [github.com/go-ldap/ldap/v3](https://github.com/go-ldap/ldap) | LDAP client |
| [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) | PDF generation |

## Building from Source

**Requirements:** Go 1.24+

```bash
git clone https://github.com/misc-lab/ldap-user-tool.git
cd ldap-user-tool
go mod tidy
go build -o ldap-user-tool .
```

### Cross-compile for the LDAP server (Linux amd64):

```bash
GOOS=linux GOARCH=amd64 go build -o ldap-user-tool .
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `cannot read /etc/ldap/admin_pass` | Missing or wrong permissions on password file | Create file and set `chmod 600` |
| `failed to connect to LDAP` | LDAP server unreachable or TLS issue | Verify `ldaps://ldap01.misc-lab.org` is reachable |
| `failed to bind as admin` | Wrong admin password | Check `/etc/ldap/admin_pass` content |
| `failed to create user` | Duplicate CN or UID | Verify user doesn't already exist |
| `group 'X' not found` | Group name doesn't exist in LDAP | Check group name spelling, list groups with `ldapsearch` |
| `sendmail failed` | Postfix not running or not installed | Run `sudo systemctl start postfix` |

## License

Internal tool for misc-lab.org.
