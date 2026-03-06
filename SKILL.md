---
name: mog
description: Microsoft Ops Gadget — CLI for Microsoft 365 (Mail, Calendar, Drive, Contacts, Tasks, Word, PowerPoint, Excel, OneNote).
---

# mog — Microsoft Ops Gadget

CLI for Microsoft 365: Mail, Calendar, OneDrive, Contacts, Tasks, Word, PowerPoint, Excel, OneNote.

The Microsoft counterpart to `gog` (Google Ops Gadget). Same patterns, different cloud.

## Quick Reference

For comprehensive usage, run:
```bash
mog --ai-help
```

This outputs the full dashdash-compliant documentation including:
- Setup/Prerequisites
- All commands and options
- Date/time formats
- Examples (positive and negative)
- Troubleshooting
- Slug system explanation
- gog compatibility notes

## Modules

| Module | Commands |
|--------|----------|
| **mail** | search, get, send, folders, drafts, attachment |
| **calendar** | list, create, get, update, delete, calendars, respond, freebusy, acl |
| **drive** | ls, search, download, upload, mkdir, move, rename, copy, rm |
| **contacts** | list, search, get, create, update, delete, directory |
| **tasks** | lists, list, add, done, undo, delete, clear |
| **word** | list, export, copy |
| **ppt** | list, export, copy |
| **excel** | list, get, update, append, create, metadata, tables, add-sheet, clear, copy, export |
| **onenote** | notebooks, sections, pages, get, create-notebook, create-section, create-page, delete, search |

## Quick Start

```bash
# Mail
mog mail search "from:someone" --max 10
mog mail send --to a@b.com --subject "Hi" --body "Hello"
mog mail send --to a@b.com --subject "Report" --body-file report.md
mog mail send --to a@b.com --subject "Newsletter" --body-html "<h1>Hello</h1>"
cat draft.txt | mog mail send --to a@b.com --subject "Hi" --body-file -

# Calendar
mog calendar list
mog calendar create --summary "Meeting" --from 2025-01-15T10:00:00 --to 2025-01-15T11:00:00
mog calendar freebusy alice@example.com bob@example.com

# Drive
mog drive ls
mog drive upload ./file.pdf
mog drive download <slug> --out ./file.pdf

# Tasks
mog tasks list
mog tasks add "Buy milk" --due tomorrow
mog tasks clear

# Contacts
mog contacts list
mog contacts directory "john"

# Excel
mog excel list
mog excel get <id> Sheet1 A1:D10
mog excel update <id> Sheet1 A1:B2 val1 val2 val3 val4
mog excel append <id> TableName col1 col2 col3

# OneNote
mog onenote notebooks
mog onenote search "meeting notes"
```

## Slugs

mog generates 8-character slugs for Microsoft's long GUIDs:
- `a3f2c891` instead of `AQMkADAwATMzAGZmAS04MDViLTRiNzgt...`
- All commands accept slugs or full IDs
- Use `--verbose` to see full IDs

## Command Aliases

- `mog cal` → `mog calendar`
- `mog todo` → `mog tasks`

## Named Aliases

Create memorable `@names` for frequently-used IDs:

```bash
mog alias set @standup f1a2b3c4
mog calendar get @standup
mog alias list
mog alias rm @standup
```

- `@` prefix, shell-safe
- Chain resolution: `@alias` → slug → full ID
- Stored per-account in `~/.config/mog/{account}/aliases.json`

## Multi-Account Support

Manage multiple Microsoft 365 accounts (personal, work, etc.):

```bash
# Setup additional accounts
mog auth login --client-id YOUR_CLIENT_ID --account work
mog auth login --client-id YOUR_CLIENT_ID --account personal

# Use specific account
mog mail search "*" --account work
mog calendar list --account personal

# Or set via environment variable
export MOG_ACCOUNT=work
mog mail search "*"

# List configured accounts
mog auth list
```

## Credential Storage

OAuth tokens stored in config directory (0600 permissions):

| Platform | Location |
|----------|----------|
| **macOS** | `~/.config/mog/{account}/` |
| **Linux** | `~/.config/mog/{account}/` |
| **Windows** | `%USERPROFILE%\.config\mog\{account}\` |

Files per account:
- `tokens.json` - OAuth tokens
- `settings.json` - Client ID
- `slugs.json` - Slug cache

Default account is `default`. Existing single-account setups auto-migrate.

## See Also

- `mog --ai-help` - Full documentation
- `mog <command> --help` - Command-specific help
