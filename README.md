# mog — Microsoft Ops Gadget

CLI for Microsoft 365: Mail, Calendar, OneDrive, Contacts, Tasks, Word, PowerPoint, Excel, OneNote.

The Microsoft counterpart to `gog` (Google Ops Gadget). Same patterns, different cloud.

## Installation

```bash
go install github.com/visionik/mog/cmd/mog@latest
```

Or build from source:

```bash
git clone https://github.com/visionik/mog.git
cd mog
go install ./cmd/mog
```

## Setup

### 1. Create an Azure AD App Registration

1. Go to [Azure Portal](https://portal.azure.com) → Azure Active Directory → App registrations
2. Click "New registration"
3. Name: "mog CLI" (or whatever you prefer)
4. Supported account types: "Accounts in any organizational directory and personal Microsoft accounts"
5. Redirect URI: Leave blank (we use device code flow)
6. Click "Register"
7. Copy the "Application (client) ID"

### 2. Configure API Permissions

In your app registration:
1. Go to "API permissions"
2. Add these Microsoft Graph delegated permissions:
   - `User.Read`
   - `offline_access`
   - `Mail.ReadWrite`
   - `Mail.Send`
   - `Calendars.ReadWrite`
   - `Files.ReadWrite.All`
   - `Contacts.ReadWrite`
   - `Tasks.ReadWrite`
   - `Notes.ReadWrite`

### 3. Enable Public Client Flow

1. Go to "Authentication"
2. Under "Advanced settings", set "Allow public client flows" to "Yes"
3. Save

### 4. Login

```bash
mog auth login --client-id YOUR_CLIENT_ID
```

Follow the device code flow instructions.

## Quick Start

```bash
# Check auth status
mog auth status

# Mail
mog mail search "*" --max 10
mog mail get <id>
mog mail send --to user@example.com --subject "Hello" --body "Hi!"

# Calendar
mog calendar list
mog calendar create --summary "Meeting" --from 2025-01-15T10:00:00 --to 2025-01-15T11:00:00
mog calendar respond <eventId> accept

# Drive (OneDrive)
mog drive ls
mog drive upload ./file.pdf
mog drive download <id> --out ./downloaded.pdf

# Contacts
mog contacts list
mog contacts search "john"
mog contacts directory "jane"  # Search org directory

# Tasks (Microsoft To-Do)
mog tasks lists
mog tasks list
mog tasks add "Buy milk" --due tomorrow --important
mog tasks done <taskId> --list <listId>

# OneNote
mog onenote notebooks
mog onenote sections <notebookId>
mog onenote pages <sectionId>
mog onenote get <pageId>
```

## Slugs

Microsoft Graph uses extremely long IDs. mog generates 8-character slugs:

- `a3f2c891` instead of `AQMkADAwATMzAGZmAS04MDViLTRiNzgt...`
- All commands accept slugs or full IDs
- Use `--verbose` to see full IDs
- Slugs are cached in `~/.config/mog/slugs.json`

## Global Flags

```
--json           JSON output (for scripting)
--plain          Plain text output (TSV)
--verbose, -v    Show full IDs
--force          Skip confirmations
--no-input       Never prompt (CI mode)
--ai-help        Detailed help for AI/LLM agents
```

## For AI Agents

Run `mog --ai-help` for comprehensive documentation suitable for LLM consumption.

## Modules

| Module | Description |
|--------|-------------|
| `mail` | Outlook mail |
| `calendar` | Outlook calendar |
| `drive` | OneDrive |
| `contacts` | People/contacts |
| `tasks` | Microsoft To-Do |
| `onenote` | OneNote |
| `word` | Word documents (basic) |
| `ppt` | PowerPoint (basic) |
| `excel` | Excel workbooks (basic) |

## Aliases

- `mog cal` → `mog calendar`
- `mog todo` → `mog tasks`
- `mog email` → `mog mail`

## Configuration

| File | Purpose |
|------|---------|
| `~/.config/mog/settings.json` | Client ID |
| `~/.config/mog/tokens.json` | OAuth tokens |
| `~/.config/mog/slugs.json` | ID slug cache |

## License

MIT
