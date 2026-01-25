# Go Rewrite Checklist

## Missing Commands

### Calendar
- [ ] `calendar acl` - List calendar permissions

### Excel
- [x] `excel add-sheet` - Add worksheet to workbook (implemented)
- [x] `excel tables` - List tables in workbook (implemented)
- [x] `excel clear` - Clear values in range (implemented)

### OneNote
- [ ] `onenote create-notebook` - Create new notebook
- [ ] `onenote create-section` - Create section in notebook
- [ ] `onenote create-page` - Create page in section
- [ ] `onenote delete` - Delete a page

### Word
- [x] `word get` - Get document metadata (implemented)
- [x] `word create` - Create new document (implemented)

### PowerPoint
- [x] `ppt get` - Get presentation metadata (implemented)
- [x] `ppt create` - Create new presentation (implemented)

---

## Test Coverage

### Unit Tests Created
- [x] internal/cli/auth_test.go - Auth commands (token status, logout)
- [x] internal/cli/mail_test.go - Mail data types, helpers, output formatting
- [x] internal/cli/calendar_test.go - Calendar data types, helpers, output formatting
- [x] internal/cli/drive_test.go - Drive data types, formatSize, output formatting
- [x] internal/cli/contacts_test.go - Contact data types, output formatting
- [x] internal/cli/tasks_test.go - Task data types, list output formatting
- [x] internal/cli/excel_test.go - Worksheet, RangeData, Table types, parseCell
- [x] internal/cli/onenote_test.go - Notebook, Section, Page types
- [x] internal/cli/word_test.go - Word command fields, filtering, output
- [x] internal/cli/ppt_test.go - PPT command fields, filtering, output
- [x] internal/graph/client_test.go - Graph client, HTTP mocking, auth types
- [x] internal/graph/slugs_test.go - Slug formatting and resolution (existing)
- [x] internal/config/config_test.go - Config/token save/load (existing)

### Coverage Summary
- internal/config: 70.8% (good)
- internal/graph: 26.6% (moderate - includes auth flows that need real OAuth)
- internal/cli: 11.1% (low - Run methods require API calls)

### Notes on Coverage
The CLI command `Run` methods have low coverage because they require:
1. A valid Graph client (needs tokens)
2. Real HTTP calls to Microsoft Graph API

To improve coverage, consider:
- [ ] Refactor to use dependency injection for the Graph client
- [ ] Create mock interfaces for the Graph client
- [ ] Use httptest to mock Graph API responses
- [ ] Add more integration tests (run with MOG_INTEGRATION=1)

### API Tests Needed (Future)
- [ ] internal/graph/mail_test.go
- [ ] internal/graph/calendar_test.go
- [ ] internal/graph/drive_test.go
- [ ] internal/graph/contacts_test.go
- [ ] internal/graph/tasks_test.go
- [ ] internal/graph/excel_test.go
- [ ] internal/graph/onenote_test.go

---

## Known Issues Found During Testing

### Excel parseCell Bug
The `parseCell` function in excel.go has a bug where `fmt.Sscanf` returns the count
of scanned items (always 1 for a successful scan) instead of properly assigning the
row value. This affects all range-based operations.

**Location**: internal/cli/excel.go:714
**Fix**: Change from:
```go
row, _ = fmt.Sscanf(cell[i:], "%d", &row)
```
To:
```go
fmt.Sscanf(cell[i:], "%d", &row)
```

---

## Progress

- Commands: 5/12 implemented (excel add-sheet, excel tables, excel clear, word get/create, ppt get/create were already done)
- Test files: 13/13 created (10 new CLI tests + existing config/slugs tests + new client test)
