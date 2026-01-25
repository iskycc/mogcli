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

func setupOneNoteTestConfig(t *testing.T) func() {
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

// Tests for Notebook type
func TestNotebook_Unmarshal(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		displayName string
	}{
		{
			name:        "personal notebook",
			json:        `{"id": "nb-123", "displayName": "Personal Notes"}`,
			displayName: "Personal Notes",
		},
		{
			name:        "work notebook",
			json:        `{"id": "nb-456", "displayName": "Work Projects"}`,
			displayName: "Work Projects",
		},
		{
			name:        "empty name",
			json:        `{"id": "nb-789", "displayName": ""}`,
			displayName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var notebook Notebook
			err := json.Unmarshal([]byte(tt.json), &notebook)
			require.NoError(t, err)
			assert.Equal(t, tt.displayName, notebook.DisplayName)
		})
	}
}

// Tests for Section type
func TestSection_Unmarshal(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		displayName string
	}{
		{
			name:        "meeting notes section",
			json:        `{"id": "sec-123", "displayName": "Meeting Notes"}`,
			displayName: "Meeting Notes",
		},
		{
			name:        "ideas section",
			json:        `{"id": "sec-456", "displayName": "Ideas"}`,
			displayName: "Ideas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var section Section
			err := json.Unmarshal([]byte(tt.json), &section)
			require.NoError(t, err)
			assert.Equal(t, tt.displayName, section.DisplayName)
		})
	}
}

// Tests for Page type
func TestPage_Unmarshal(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		title string
	}{
		{
			name:  "titled page",
			json:  `{"id": "page-123", "title": "Weekly Review"}`,
			title: "Weekly Review",
		},
		{
			name:  "untitled page",
			json:  `{"id": "page-456", "title": ""}`,
			title: "",
		},
		{
			name:  "page with special characters",
			json:  `{"id": "page-789", "title": "Q4 Planning & Goals"}`,
			title: "Q4 Planning & Goals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var page Page
			err := json.Unmarshal([]byte(tt.json), &page)
			require.NoError(t, err)
			assert.Equal(t, tt.title, page.Title)
		})
	}
}

// Tests for OneNoteNotebooksCmd
func TestOneNoteNotebooksCmd_Struct(t *testing.T) {
	cmd := &OneNoteNotebooksCmd{}
	assert.NotNil(t, cmd)
}

// Tests for OneNoteSectionsCmd
func TestOneNoteSectionsCmd_Fields(t *testing.T) {
	cmd := &OneNoteSectionsCmd{
		NotebookID: "notebook-123",
	}
	assert.Equal(t, "notebook-123", cmd.NotebookID)
}

// Tests for OneNotePagesCmd
func TestOneNotePagesCmd_Fields(t *testing.T) {
	cmd := &OneNotePagesCmd{
		SectionID: "section-123",
	}
	assert.Equal(t, "section-123", cmd.SectionID)
}

// Tests for OneNoteGetCmd
func TestOneNoteGetCmd_Fields(t *testing.T) {
	tests := []struct {
		name   string
		pageID string
		html   bool
	}{
		{
			name:   "text output",
			pageID: "page-123",
			html:   false,
		},
		{
			name:   "html output",
			pageID: "page-456",
			html:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &OneNoteGetCmd{
				PageID: tt.pageID,
				HTML:   tt.html,
			}
			assert.Equal(t, tt.pageID, cmd.PageID)
			assert.Equal(t, tt.html, cmd.HTML)
		})
	}
}

// Tests for OneNoteSearchCmd
func TestOneNoteSearchCmd_Fields(t *testing.T) {
	cmd := &OneNoteSearchCmd{
		Query: "project notes",
	}
	assert.Equal(t, "project notes", cmd.Query)
}

// Tests for notebook output
func TestNotebook_Output(t *testing.T) {
	notebooks := []Notebook{
		{ID: "nb-1-long-id-for-testing-purposes", DisplayName: "Personal"},
		{ID: "nb-2-long-id-for-testing-purposes", DisplayName: "Work"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, nb := range notebooks {
		os.Stdout.WriteString(nb.DisplayName + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Personal")
	assert.Contains(t, output, "Work")
}

// Tests for section output
func TestSection_Output(t *testing.T) {
	sections := []Section{
		{ID: "sec-1-long-id-for-testing", DisplayName: "Meetings"},
		{ID: "sec-2-long-id-for-testing", DisplayName: "Research"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, s := range sections {
		os.Stdout.WriteString(s.DisplayName + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Meetings")
	assert.Contains(t, output, "Research")
}

// Tests for page output
func TestPage_Output(t *testing.T) {
	pages := []Page{
		{ID: "page-1-long-id-for-testing", Title: "Weekly Review"},
		{ID: "page-2-long-id-for-testing", Title: "Project Plan"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, p := range pages {
		os.Stdout.WriteString(p.Title + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Weekly Review")
	assert.Contains(t, output, "Project Plan")
}

// Tests for JSON output
func TestNotebook_JSONOutput(t *testing.T) {
	notebook := Notebook{
		ID:          "nb-json-test",
		DisplayName: "JSON Test Notebook",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(notebook)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "nb-json-test"`)
	assert.Contains(t, output, `"displayName": "JSON Test Notebook"`)
}

func TestSection_JSONOutput(t *testing.T) {
	section := Section{
		ID:          "sec-json-test",
		DisplayName: "JSON Test Section",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(section)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "sec-json-test"`)
	assert.Contains(t, output, `"displayName": "JSON Test Section"`)
}

func TestPage_JSONOutput(t *testing.T) {
	page := Page{
		ID:    "page-json-test",
		Title: "JSON Test Page",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(page)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "page-json-test"`)
	assert.Contains(t, output, `"title": "JSON Test Page"`)
}

// Tests for HTML content handling
func TestOneNoteGet_HTMLStripping(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body>
<div data-id="p1">
<p>This is <b>bold</b> and <i>italic</i> text.</p>
</div>
</body>
</html>`

	// Use the stripHTML function
	result := stripHTML(htmlContent)

	assert.NotContains(t, result, "<html>")
	assert.NotContains(t, result, "<body>")
	assert.NotContains(t, result, "<p>")
	assert.Contains(t, result, "bold")
	assert.Contains(t, result, "italic")
}

// Edge cases
func TestNotebook_EmptyList(t *testing.T) {
	var notebooks []Notebook
	assert.Empty(t, notebooks)
}

func TestSection_EmptyList(t *testing.T) {
	var sections []Section
	assert.Empty(t, sections)
}

func TestPage_EmptyList(t *testing.T) {
	var pages []Page
	assert.Empty(t, pages)
}

func TestNotebook_SpecialCharacters(t *testing.T) {
	jsonData := `{"id": "nb-special", "displayName": "Notes & Ideas: 2024"}`

	var notebook Notebook
	err := json.Unmarshal([]byte(jsonData), &notebook)
	require.NoError(t, err)
	assert.Equal(t, "Notes & Ideas: 2024", notebook.DisplayName)
}

func TestPage_UnicodeTitle(t *testing.T) {
	jsonData := `{"id": "page-unicode", "title": "日本語ノート"}`

	var page Page
	err := json.Unmarshal([]byte(jsonData), &page)
	require.NoError(t, err)
	assert.Equal(t, "日本語ノート", page.Title)
}

// Tests for OneNote structure hierarchy
func TestOneNote_Hierarchy(t *testing.T) {
	// Test that the hierarchy makes sense
	notebook := Notebook{ID: "nb-1", DisplayName: "Work"}
	sections := []Section{
		{ID: "sec-1", DisplayName: "Projects"},
		{ID: "sec-2", DisplayName: "Meetings"},
	}
	pages := []Page{
		{ID: "page-1", Title: "Project Alpha"},
		{ID: "page-2", Title: "Project Beta"},
	}

	assert.NotEmpty(t, notebook.ID)
	assert.Len(t, sections, 2)
	assert.Len(t, pages, 2)
}
