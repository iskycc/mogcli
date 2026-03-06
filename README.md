# 📊 mog — Microsoft Ops Gadget

> **CLI for Microsoft 365** — Mail, Calendar, Drive, Contacts, Tasks, Word, PowerPoint, Excel, OneNote

[![Go Reference](https://pkg.go.dev/badge/github.com/visionik/mogcli.svg)](https://pkg.go.dev/github.com/visionik/mogcli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

The **Microsoft** counterpart to [gog](https://github.com/visionik/gog) (Google Ops Gadget). Same patterns, different cloud.

---

## ✨ Features

| Module | Description |
|--------|-------------|
| 📧 **Mail** | Search, send, drafts, attachments, folders |
| 📅 **Calendar** | Events, create, respond, freebusy, ACL |
| 📁 **Drive** | OneDrive files — list, upload, download, move |
| 👥 **Contacts** | Personal contacts + org directory lookup |
| ✅ **Tasks** | Microsoft To-Do — lists, add, complete, clear |
| 📝 **Word** | Documents — list, export, copy |
| 📊 **PowerPoint** | Presentations — list, export, copy |
| 📈 **Excel** | Spreadsheets — read, write, tables, export |
| 📓 **OneNote** | Notebooks, sections, pages, search |

**Extras:**
- 🔗 **Slug system** — 8-char shorthand for Microsoft's long GUIDs
- 🤖 **AI-friendly** — `--ai-help` outputs comprehensive docs for LLMs
- 🔄 **gog-compatible** — Same flags and patterns for muscle memory

---

## 🚀 Quick Start

```bash
# Install
go install github.com/visionik/mogcli/cmd/mog@latest

# Authenticate (see Setup below for Azure AD app)
mog auth login --client-id YOUR_CLIENT_ID

# Check mail
mog mail search "*" --max 10

# Send email
mog mail send --to bob@example.com --subject "Hello" --body "Hi Bob!"

# List calendar events
mog calendar list

# Create event with attendees
mog calendar create --summary "Meeting" \
  --from 2025-01-15T10:00:00 --to 2025-01-15T11:00:00 \
  --attendees "alice@example.com"

# Upload to OneDrive
mog drive upload ./report.pdf

# Add a task
mog tasks add "Review PR" --due tomorrow --important

# Read Excel spreadsheet
mog excel get myworkbook.xlsx Sheet1 A1:D10

# Search OneNote
mog onenote search "meeting notes"
```

---

## 📦 Installation

```bash
# Go install (recommended)
go install github.com/visionik/mogcli/cmd/mog@latest

# Or clone for development
git clone https://github.com/visionik/mogcli.git
cd mogcli
go build -o mog ./cmd/mog
```

---

## ⚙️ Setup — Azure AD App

### 1. Create App Registration

1. Go to [Azure Portal](https://portal.azure.com) → **App registrations** → **New registration**
2. **Name:** `mog CLI` (or any name)
3. **Supported account types:** Select based on your needs
4. **Redirect URI:** Leave blank (uses device code flow)

### 2. Add API Permissions

Add these **Delegated** permissions:

| Permission | Description |
|------------|-------------|
| `User.Read` | Sign in and read user profile |
| `offline_access` | Maintain access (refresh tokens) |
| `Mail.ReadWrite` | Read and write mail |
| `Mail.Send` | Send mail |
| `Calendars.ReadWrite` | Full calendar access |
| `Files.ReadWrite.All` | Full OneDrive access |
| `Contacts.Read` | Read contacts |
| `Contacts.ReadWrite` | Full contacts access |
| `People.Read` | Read people |
| `Tasks.ReadWrite` | Read and write tasks |
| `Notes.ReadWrite` | Read and write OneNote |

### 3. Authenticate

```bash
mog auth login --client-id YOUR_CLIENT_ID
```

Opens a browser for Microsoft login. Tokens stored at `~/.config/mog/tokens.json`.

### 4. Verify

```bash
mog auth status
```

---

## 👥 Multi-Account Support

Manage multiple Microsoft 365 accounts (personal, work, etc.):

### Setup Additional Accounts

```bash
# Login with a named account
mog auth login --client-id YOUR_CLIENT_ID --account work
mog auth login --client-id YOUR_CLIENT_ID --account personal
```

### Using Accounts

```bash
# Specify account per-command
mog mail search "*" --account work
mog calendar list --account personal

# Or set via environment variable
export MOG_ACCOUNT=work
mog mail search "*"
```

### List Configured Accounts

```bash
mog auth list
```

### Account Storage

Each account has isolated storage:
```
~/.config/mog/
  default/
    settings.json
    tokens.json
    slugs.json
  work/
    settings.json
    tokens.json
    slugs.json
```

Existing single-account setups are auto-migrated to `default`.

---

## 📖 Command Reference

### Global Options

| Option | Description |
|--------|-------------|
| `--account`, `-a` | Account name (default: "default", env: `MOG_ACCOUNT`) |
| `--json` | Output JSON (best for scripting) |
| `--plain` | Stable text output (TSV, no colors) |
| `--verbose` | Show full IDs and extra details |
| `--force` | Skip confirmations |
| `--no-input` | Never prompt (CI mode) |
| `--ai-help` | Full docs for AI agents |

---

### 📧 Mail

```bash
mog mail search <query>              # Search messages
mog mail search "*" --max 10         # Recent messages
mog mail get <id>                    # Read a message
mog mail send --to X --subject Y --body Z
mog mail folders                     # List folders

# Drafts
mog mail drafts list
mog mail drafts create --to X --subject Y --body Z
mog mail drafts send <draftId>

# Attachments
mog mail attachment list <messageId>
mog mail attachment download <messageId> <attachmentId> --out ./file.pdf
```

---

### 📅 Calendar

```bash
mog calendar list                    # Upcoming events
mog calendar list --from 2025-01-01 --to 2025-01-31
mog calendar calendars               # List calendars

mog calendar create --summary "Meeting" \
  --from 2025-01-15T10:00:00 \
  --to 2025-01-15T11:00:00

mog calendar get <eventId>
mog calendar update <eventId> --summary "New Title"
mog calendar delete <eventId>

# Respond to invites
mog calendar respond <eventId> accept
mog calendar respond <eventId> decline --comment "Can't make it"

# Check availability
mog calendar freebusy alice@example.com bob@example.com \
  --start 2025-01-15T09:00:00 --end 2025-01-15T17:00:00

# View permissions
mog calendar acl
```

**Alias:** `mog cal` → `mog calendar`

---

### 📁 Drive (OneDrive)

```bash
mog drive ls                         # Root folder
mog drive ls /Documents              # Specific path
mog drive search "report"            # Search files

mog drive download <id> --out ./file.pdf
mog drive upload ./doc.pdf
mog drive upload ./doc.pdf --folder <folderId> --name "renamed.pdf"

mog drive mkdir "New Folder"
mog drive move <id> <destinationId>
mog drive rename <id> "new-name.pdf"
mog drive copy <id> --name "copy.pdf"
mog drive rm <id>
```

---

### ✅ Tasks (Microsoft To-Do)

```bash
mog tasks lists                      # List task lists
mog tasks list                       # Tasks in default list
mog tasks list <listId>              # Tasks in specific list
mog tasks list --all                 # Include completed

mog tasks add "Buy milk"
mog tasks add "Call mom" --due tomorrow --notes "Birthday"
mog tasks add "Review PR" --list Work --due monday --important

mog tasks done <taskId>
mog tasks undo <taskId>
mog tasks delete <taskId>
mog tasks clear                      # Clear completed tasks
mog tasks clear <listId>             # Clear from specific list
```

**Alias:** `mog todo` → `mog tasks`

---

### 👥 Contacts

```bash
mog contacts list
mog contacts search "john"
mog contacts get <id>

mog contacts create --name "John Doe" --email "john@example.com"
mog contacts update <id> --email "new@example.com"
mog contacts delete <id>

mog contacts directory "john"        # Org directory lookup
```

---

### 📈 Excel

```bash
mog excel list                       # List workbooks
mog excel metadata <id>              # List worksheets

# Read data
mog excel get <id>                   # First sheet, used range
mog excel get <id> Sheet1 A1:D10     # Specific range

# Write data (positional values fill row by row)
mog excel update <id> Sheet1 A1:B2 val1 val2 val3 val4

# Append to table
mog excel append <id> TableName col1 col2 col3

# Create & manage
mog excel create "Budget 2025"
mog excel add-sheet <id> --name "Q2"
mog excel tables <id>
mog excel clear <id> Sheet1 A1:C10   # Clear values (keep formatting)
mog excel copy <id> "Budget Copy"

# Export
mog excel export <id> --out ./data.xlsx
mog excel export <id> --format csv --out ./data.csv
```

---

### 📓 OneNote

```bash
mog onenote notebooks                # List notebooks
mog onenote sections <notebookId>    # List sections
mog onenote pages <sectionId>        # List pages
mog onenote get <pageId>             # Get page content (text)
mog onenote get <pageId> --html      # Get raw HTML

mog onenote create-notebook "Work Notes"
mog onenote create-section <notebookId> "January"
mog onenote create-page <sectionId> "Meeting Notes" "Content here"

mog onenote delete <pageId>
mog onenote search "meeting"
```

---

### 📝 Word

```bash
mog word list                        # List documents
mog word export <id> --out ./doc.docx
mog word export <id> --format pdf --out ./doc.pdf
mog word copy <id> "Copy of Report"
```

---

### 📊 PowerPoint

```bash
mog ppt list                         # List presentations
mog ppt export <id> --out ./deck.pptx
mog ppt export <id> --format pdf --out ./deck.pdf
mog ppt copy <id> "Copy of Deck"
```

---

## 🔗 Slug System

Microsoft Graph uses very long GUIDs (100+ characters). mog generates 8-character slugs:

```
Full:  AQMkADAwATMzAGZmAS04MDViLTRiNzgtMDA...
Slug:  a3f2c891
```

- ✅ All commands output slugs by default
- ✅ All commands accept slugs or full IDs
- ✅ Use `--verbose` to also see full IDs
- ✅ Slugs cached in `~/.config/mog/slugs.json`
- ✅ `mog auth logout` clears the cache

## 🏷️ Aliases

Create memorable names for frequently-used IDs or slugs:

```bash
# Create aliases
mog alias set @standup f1a2b3c4
mog alias set @budget "AQMkADAwATMz..."

# Use aliases anywhere you'd use an ID
mog calendar get @standup
mog excel get @budget Sheet1 A1:D10

# Manage aliases
mog alias list
mog alias get @standup
mog alias rm @standup
```

- ✅ `@` prefix for aliases (shell-safe, no quoting needed)
- ✅ Aliases resolve through slugs: `@standup` → `f1a2b3c4` → full ID
- ✅ Stored per-account in `~/.config/mog/{account}/aliases.json`

---

## 🤖 AI-Friendly

Run `mog --ai-help` for comprehensive documentation including:

- All commands with options
- Date/time format specifications
- Positive and negative examples
- Exit codes and piping patterns
- Troubleshooting guide

Follows the [dashdash](https://github.com/visionik/dashdash) specification.

---

## 🔄 gog Compatibility

mog follows [gog](https://github.com/visionik/gog) patterns for muscle memory across clouds:

| Pattern | mog | gog |
|---------|-----|-----|
| Calendar events | `--summary`, `--from`, `--to` | Same |
| Task notes | `--notes` | Same |
| Output format | `--json`, `--plain` | Same |
| Max results | `--max` | Same |
| Excel read | `mog excel get <id> Sheet1 A1:D10` | `gog sheets get <id> Sheet1!A1:D10` |
| Spreadsheet write | `mog excel update <id> ...` | `gog sheets update <id> ...` |

---

## 🗂️ Configuration

| File | Purpose |
|------|---------|
| `~/.config/mog/tokens.json` | OAuth tokens (sensitive) |
| `~/.config/mog/settings.json` | Client ID and settings |
| `~/.config/mog/slugs.json` | ID-to-slug cache |

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `MOG_CLIENT_ID` | Azure AD client ID (alternative to --client-id) |

---

## 🛠️ Development

```bash
# Using Taskfile (recommended)
task test              # Run tests
task test:coverage     # With coverage
task lint              # Lint
task fmt               # Format
task check             # All checks

# Or directly with Go
go test ./...
go build ./cmd/mog
```

---

## 📄 License

MIT

---

## 👨‍💻 Developed By

**[visionik](mailto:visionik@pobox.com)** and **Vinston 🐺** ([Clawdbot](https://github.com/clawdbot/clawdbot)) using the visionik.md framework.
