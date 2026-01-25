# Go Rewrite Checklist

## Missing Commands

### Calendar
- [x] `calendar acl` - List calendar permissions

### Excel
- [x] `excel add-sheet` - Add worksheet to workbook (already implemented)
- [x] `excel tables` - List tables in workbook (already implemented)
- [x] `excel clear` - Clear values in range (already implemented)

### OneNote
- [x] `onenote create-notebook` - Create new notebook
- [x] `onenote create-section` - Create section in notebook
- [x] `onenote create-page` - Create page in section
- [x] `onenote delete` - Delete a page

### Word
- [x] `word get` - Get document metadata (already implemented)
- [x] `word create` - Create new document (already implemented)

### PowerPoint
- [x] `ppt get` - Get presentation metadata (already implemented)
- [x] `ppt create` - Create new presentation (already implemented)

---

## Test Coverage Gaps

### Unit Tests Needed
- [ ] internal/cli/mail_test.go
- [ ] internal/cli/calendar_test.go
- [ ] internal/cli/drive_test.go
- [ ] internal/cli/contacts_test.go
- [ ] internal/cli/tasks_test.go
- [ ] internal/cli/excel_test.go
- [ ] internal/cli/onenote_test.go
- [ ] internal/cli/word_test.go
- [ ] internal/cli/ppt_test.go
- [ ] internal/cli/auth_test.go

### API Tests Needed
- [ ] internal/graph/mail_test.go
- [ ] internal/graph/calendar_test.go
- [ ] internal/graph/drive_test.go
- [ ] internal/graph/contacts_test.go
- [ ] internal/graph/tasks_test.go
- [ ] internal/graph/excel_test.go
- [ ] internal/graph/onenote_test.go

---

## Progress

- Commands: 12/12 implemented ✅
- Test files: 0/17 created
