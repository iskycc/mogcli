// Package integration provides integration tests for CLI commands.
// These tests require valid authentication and make real API calls.
//
// Run with: MOG_INTEGRATION=1 go test -v ./internal/cli/integration/...
// Or: task test:integration
package integration

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoIntegration(t *testing.T) {
	if os.Getenv("MOG_INTEGRATION") == "" {
		t.Skip("Skipping integration test (set MOG_INTEGRATION=1 to run)")
	}
}

func runMog(t *testing.T, args ...string) (string, string, error) {
	// Use the installed binary or build it
	binary := os.Getenv("MOG_BINARY")
	if binary == "" {
		binary = "mogcli"
	}

	cmd := exec.Command(binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestIntegration_AuthStatus(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "auth", "status")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Status:")
}

func TestIntegration_MailFolders(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "mail", "folders")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Inbox")
}

func TestIntegration_MailFoldersJSON(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "mail", "folders", "--json")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(stdout), "["))
}

func TestIntegration_MailSearch(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "mail", "search", "*", "--max", "1")
	require.NoError(t, err)
	// Should have at least a header or message
	assert.NotEmpty(t, stdout)
}

func TestIntegration_CalendarCalendars(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "calendar", "calendars")
	require.NoError(t, err)
	// Should list at least one calendar
	assert.NotEmpty(t, stdout)
}

func TestIntegration_CalendarList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "calendar", "list", "--max", "1")
	// May or may not have events, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_DriveLs(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "drive", "ls")
	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
}

func TestIntegration_DriveLsJSON(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "drive", "ls", "--json")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(stdout), "["))
}

func TestIntegration_ContactsList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "contacts", "list", "--max", "1")
	// May have no contacts, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_TasksLists(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "tasks", "lists")
	require.NoError(t, err)
	// Should have at least one task list
	assert.NotEmpty(t, stdout)
}

func TestIntegration_TasksList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "tasks", "list")
	// May have no tasks, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_OneNoteNotebooks(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "onenote", "notebooks")
	// May have no notebooks, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_ExcelList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "excel", "list", "--max", "1")
	// May have no workbooks, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_WordList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "word", "list", "--max", "1")
	// May have no documents, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_PPTList(t *testing.T) {
	skipIfNoIntegration(t)

	_, _, err := runMog(t, "ppt", "list", "--max", "1")
	// May have no presentations, but shouldn't error
	assert.NoError(t, err)
}

func TestIntegration_AIHelp(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "--ai-help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "mog")
	assert.Contains(t, stdout, "Microsoft")
}

func TestIntegration_Help(t *testing.T) {
	skipIfNoIntegration(t)

	stdout, _, err := runMog(t, "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "mail")
	assert.Contains(t, stdout, "calendar")
}
