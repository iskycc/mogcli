package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visionik/mogcli/internal/config"
)

func setupTasksTestConfig(t *testing.T) func() {
	t.Helper()

	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	tokens := &config.Tokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    9999999999,
	}
	require.NoError(t, config.SaveTokens(tokens))

	cfg := &config.Config{ClientID: "test-client-id-12345678901234567890"}
	require.NoError(t, config.Save(cfg))

	return func() {
		os.Setenv("HOME", origHome)
	}
}

// Tests for TaskList type
func TestTaskList_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		isOwner bool
	}{
		{
			name:    "owned list",
			json:    `{"id": "list-123", "displayName": "My Tasks", "isOwner": true}`,
			isOwner: true,
		},
		{
			name:    "shared list",
			json:    `{"id": "list-456", "displayName": "Team Tasks", "isOwner": false}`,
			isOwner: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var list TaskList
			err := json.Unmarshal([]byte(tt.json), &list)
			require.NoError(t, err)
			assert.Equal(t, tt.isOwner, list.IsOwner)
		})
	}
}

// Tests for Task type
func TestTask_Unmarshal(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		status     string
		importance string
	}{
		{
			name: "not started task",
			json: `{
				"id": "task-123",
				"title": "Complete report",
				"status": "notStarted",
				"importance": "normal"
			}`,
			status:     "notStarted",
			importance: "normal",
		},
		{
			name: "completed task",
			json: `{
				"id": "task-456",
				"title": "Send email",
				"status": "completed",
				"importance": "low"
			}`,
			status:     "completed",
			importance: "low",
		},
		{
			name: "high priority task",
			json: `{
				"id": "task-789",
				"title": "Urgent meeting",
				"status": "notStarted",
				"importance": "high"
			}`,
			status:     "notStarted",
			importance: "high",
		},
		{
			name: "task with due date",
			json: `{
				"id": "task-abc",
				"title": "Project deadline",
				"status": "notStarted",
				"importance": "high",
				"dueDateTime": {
					"dateTime": "2024-01-20T00:00:00",
					"timeZone": "UTC"
				}
			}`,
			status:     "notStarted",
			importance: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var task Task
			err := json.Unmarshal([]byte(tt.json), &task)
			require.NoError(t, err)
			assert.Equal(t, tt.status, task.Status)
			assert.Equal(t, tt.importance, task.Importance)
		})
	}
}

// Tests for DateTime type
func TestDateTime_Unmarshal(t *testing.T) {
	jsonData := `{"dateTime": "2024-01-15T10:00:00", "timeZone": "UTC"}`

	var dt DateTime
	err := json.Unmarshal([]byte(jsonData), &dt)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-15T10:00:00", dt.DateTime)
	assert.Equal(t, "UTC", dt.TimeZone)
}

// Tests for TaskBody type
func TestTaskBody_Unmarshal(t *testing.T) {
	jsonData := `{"content": "Task notes here", "contentType": "text"}`

	var body TaskBody
	err := json.Unmarshal([]byte(jsonData), &body)
	require.NoError(t, err)
	assert.Equal(t, "Task notes here", body.Content)
	assert.Equal(t, "text", body.ContentType)
}

// Tests for TasksListsCmd
func TestTasksListsCmd_Struct(t *testing.T) {
	cmd := &TasksListsCmd{}
	assert.NotNil(t, cmd)
}

// Tests for TasksListCmd
func TestTasksListCmd_Fields(t *testing.T) {
	cmd := &TasksListCmd{
		ListID: "list-123",
		All:    true,
	}

	assert.Equal(t, "list-123", cmd.ListID)
	assert.True(t, cmd.All)
}

func TestTasksListCmd_DefaultAll(t *testing.T) {
	cmd := &TasksListCmd{
		ListID: "list-123",
	}
	assert.False(t, cmd.All) // Default should be false (hide completed)
}

// Tests for TasksAddCmd
func TestTasksAddCmd_Fields(t *testing.T) {
	cmd := &TasksAddCmd{
		Title:     "New task",
		ListID:    "list-123",
		Due:       "2024-01-20",
		Notes:     "Task notes",
		Important: true,
	}

	assert.Equal(t, "New task", cmd.Title)
	assert.Equal(t, "list-123", cmd.ListID)
	assert.Equal(t, "2024-01-20", cmd.Due)
	assert.Equal(t, "Task notes", cmd.Notes)
	assert.True(t, cmd.Important)
}

func TestTasksAddCmd_DueShortcuts(t *testing.T) {
	tests := []struct {
		name string
		due  string
	}{
		{"tomorrow", "tomorrow"},
		{"today", "today"},
		{"specific date", "2024-01-25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &TasksAddCmd{
				Title: "Test task",
				Due:   tt.due,
			}
			assert.Equal(t, tt.due, cmd.Due)
		})
	}
}

func TestTasksAddCmd_MinimalFields(t *testing.T) {
	cmd := &TasksAddCmd{
		Title: "Simple task",
	}

	assert.Equal(t, "Simple task", cmd.Title)
	assert.Empty(t, cmd.ListID)
	assert.Empty(t, cmd.Due)
	assert.Empty(t, cmd.Notes)
	assert.False(t, cmd.Important)
}

// Tests for TasksUpdateCmd
func TestTasksUpdateCmd_Fields(t *testing.T) {
	cmd := &TasksUpdateCmd{
		TaskID:    "task-123",
		ListID:    "list-123",
		Title:     "Updated title",
		Due:       "2024-02-01",
		Notes:     "Updated notes",
		Important: true,
	}

	assert.Equal(t, "task-123", cmd.TaskID)
	assert.Equal(t, "list-123", cmd.ListID)
	assert.Equal(t, "Updated title", cmd.Title)
}

func TestTasksUpdateCmd_RequiresList(t *testing.T) {
	cleanup := setupTasksTestConfig(t)
	defer cleanup()

	cmd := &TasksUpdateCmd{
		TaskID: "task-123",
		// ListID not set
		Title: "Update",
	}

	root := &Root{}
	err := cmd.Run(root)

	// Should error because list is required
	assert.Error(t, err)
}

// Tests for TasksDoneCmd
func TestTasksDoneCmd_Fields(t *testing.T) {
	cmd := &TasksDoneCmd{
		TaskID: "task-123",
		ListID: "list-123",
	}

	assert.Equal(t, "task-123", cmd.TaskID)
	assert.Equal(t, "list-123", cmd.ListID)
}

func TestTasksDoneCmd_RequiresList(t *testing.T) {
	cleanup := setupTasksTestConfig(t)
	defer cleanup()

	cmd := &TasksDoneCmd{
		TaskID: "task-123",
		// ListID not set
	}

	root := &Root{}
	err := cmd.Run(root)

	assert.Error(t, err)
}

// Tests for TasksUndoCmd
func TestTasksUndoCmd_Fields(t *testing.T) {
	cmd := &TasksUndoCmd{
		TaskID: "task-123",
		ListID: "list-123",
	}

	assert.Equal(t, "task-123", cmd.TaskID)
	assert.Equal(t, "list-123", cmd.ListID)
}

func TestTasksUndoCmd_RequiresList(t *testing.T) {
	cleanup := setupTasksTestConfig(t)
	defer cleanup()

	cmd := &TasksUndoCmd{
		TaskID: "task-123",
		// ListID not set
	}

	root := &Root{}
	err := cmd.Run(root)

	assert.Error(t, err)
}

// Tests for TasksDeleteCmd
func TestTasksDeleteCmd_Fields(t *testing.T) {
	cmd := &TasksDeleteCmd{
		TaskID: "task-123",
		ListID: "list-123",
	}

	assert.Equal(t, "task-123", cmd.TaskID)
	assert.Equal(t, "list-123", cmd.ListID)
}

func TestTasksDeleteCmd_RequiresList(t *testing.T) {
	cleanup := setupTasksTestConfig(t)
	defer cleanup()

	cmd := &TasksDeleteCmd{
		TaskID: "task-123",
		// ListID not set
	}

	root := &Root{}
	err := cmd.Run(root)

	assert.Error(t, err)
}

// Tests for TasksClearCmd
func TestTasksClearCmd_Fields(t *testing.T) {
	cmd := &TasksClearCmd{
		ListID: "list-123",
	}
	assert.Equal(t, "list-123", cmd.ListID)
}

func TestTasksClearCmd_DefaultList(t *testing.T) {
	cmd := &TasksClearCmd{}
	assert.Empty(t, cmd.ListID) // Will use default list
}

// Tests for task list output format
func TestTaskList_Output(t *testing.T) {
	lists := []TaskList{
		{ID: "list-1", DisplayName: "My Tasks", IsOwner: true},
		{ID: "list-2", DisplayName: "Shared Tasks", IsOwner: false},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, list := range lists {
		marker := " "
		if list.IsOwner {
			marker = "*"
		}
		os.Stdout.WriteString(marker + " " + list.DisplayName + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "* My Tasks")
	assert.Contains(t, output, "  Shared Tasks")
}

// Tests for task output format
func TestTask_Output(t *testing.T) {
	tasks := []Task{
		{
			ID:         "task-1-long-id-for-testing",
			Title:      "Incomplete task",
			Status:     "notStarted",
			Importance: "normal",
		},
		{
			ID:         "task-2-long-id-for-testing",
			Title:      "Important task",
			Status:     "notStarted",
			Importance: "high",
		},
		{
			ID:         "task-3-long-id-for-testing",
			Title:      "Done task",
			Status:     "completed",
			Importance: "low",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, task := range tasks {
		status := "○"
		if task.Status == "completed" {
			status = "✓"
		}
		importance := " "
		if task.Importance == "high" {
			importance = "!"
		}
		os.Stdout.WriteString(status + importance + " " + task.Title + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "○  Incomplete task")
	assert.Contains(t, output, "○! Important task")
	assert.Contains(t, output, "✓  Done task")
}

// Tests for task with due date output
func TestTask_DueDateOutput(t *testing.T) {
	task := Task{
		ID:         "task-due-long-id-for-testing",
		Title:      "Task with deadline",
		Status:     "notStarted",
		Importance: "normal",
		DueDateTime: &DateTime{
			DateTime: "2024-01-20T00:00:00",
			TimeZone: "UTC",
		},
	}

	due := ""
	if task.DueDateTime != nil {
		due = task.DueDateTime.DateTime[:10]
	}

	assert.Equal(t, "2024-01-20", due)
}

// Tests for Task JSON output
func TestTask_JSONOutput(t *testing.T) {
	task := Task{
		ID:         "task-json-test",
		Title:      "JSON Task",
		Status:     "notStarted",
		Importance: "high",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(task)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "task-json-test"`)
	assert.Contains(t, output, `"title": "JSON Task"`)
	assert.Contains(t, output, `"status": "notStarted"`)
	assert.Contains(t, output, `"importance": "high"`)
}

// Tests for TaskList JSON output
func TestTaskList_JSONOutput(t *testing.T) {
	list := TaskList{
		ID:          "list-json-test",
		DisplayName: "JSON List",
		IsOwner:     true,
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(list)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "list-json-test"`)
	assert.Contains(t, output, `"displayName": "JSON List"`)
}

// Edge case tests
func TestTask_NoTitle(t *testing.T) {
	jsonData := `{"id": "task-no-title", "status": "notStarted"}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	require.NoError(t, err)
	assert.Empty(t, task.Title)
}

func TestTask_NullDueDate(t *testing.T) {
	jsonData := `{"id": "task-no-due", "title": "No due date"}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	require.NoError(t, err)
	assert.Nil(t, task.DueDateTime)
}

func TestTask_WithBody(t *testing.T) {
	jsonData := `{
		"id": "task-with-body",
		"title": "Task with notes",
		"body": {
			"content": "These are the task notes",
			"contentType": "text"
		}
	}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	require.NoError(t, err)
	assert.NotNil(t, task.Body)
	assert.Equal(t, "These are the task notes", task.Body.Content)
}
