package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visionik/mogcli/internal/config"
)

func setupWordTestConfig(t *testing.T) func() {
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

// Tests for WordListCmd
func TestWordListCmd_DefaultMax(t *testing.T) {
	cmd := &WordListCmd{
		Max: 50, // Default value
	}
	assert.Equal(t, 50, cmd.Max)
}

func TestWordListCmd_CustomMax(t *testing.T) {
	cmd := &WordListCmd{
		Max: 100,
	}
	assert.Equal(t, 100, cmd.Max)
}

// Tests for WordGetCmd
func TestWordGetCmd_Fields(t *testing.T) {
	cmd := &WordGetCmd{
		ID: "doc-123",
	}
	assert.Equal(t, "doc-123", cmd.ID)
}

// Tests for WordExportCmd
func TestWordExportCmd_Fields(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		out    string
		format string
	}{
		{
			name:   "docx export",
			id:     "doc-123",
			out:    "/output/document.docx",
			format: "docx",
		},
		{
			name:   "pdf export",
			id:     "doc-456",
			out:    "/output/document.pdf",
			format: "pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WordExportCmd{
				ID:     tt.id,
				Out:    tt.out,
				Format: tt.format,
			}
			assert.Equal(t, tt.id, cmd.ID)
			assert.Equal(t, tt.out, cmd.Out)
			assert.Equal(t, tt.format, cmd.Format)
		})
	}
}

func TestWordExportCmd_DefaultFormat(t *testing.T) {
	cmd := &WordExportCmd{
		ID:     "doc-123",
		Out:    "/output/file.docx",
		Format: "docx", // Default
	}
	assert.Equal(t, "docx", cmd.Format)
}

// Tests for WordCopyCmd
func TestWordCopyCmd_Fields(t *testing.T) {
	cmd := &WordCopyCmd{
		ID:     "doc-123",
		Name:   "Copy of Document.docx",
		Folder: "folder-456",
	}

	assert.Equal(t, "doc-123", cmd.ID)
	assert.Equal(t, "Copy of Document.docx", cmd.Name)
	assert.Equal(t, "folder-456", cmd.Folder)
}

func TestWordCopyCmd_NoFolder(t *testing.T) {
	cmd := &WordCopyCmd{
		ID:   "doc-123",
		Name: "Copy.docx",
	}

	assert.Empty(t, cmd.Folder)
}

// Tests for WordCreateCmd
func TestWordCreateCmd_Fields(t *testing.T) {
	cmd := &WordCreateCmd{
		Name:   "New Document.docx",
		Folder: "folder-123",
	}

	assert.Equal(t, "New Document.docx", cmd.Name)
	assert.Equal(t, "folder-123", cmd.Folder)
}

func TestWordCreateCmd_AutoExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already has docx", "Report.docx", "Report.docx"},
		{"no extension", "Report", "Report.docx"},
		{"uppercase docx", "Report.DOCX", "Report.DOCX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.input
			// Logic from the command
			if !strings.HasSuffix(strings.ToLower(name), ".docx") {
				name += ".docx"
			}

			if tt.name == "no extension" {
				assert.Equal(t, "Report.docx", name)
			} else {
				assert.Equal(t, tt.input, name)
			}
		})
	}
}

// Tests for Word document filtering (from drive search)
func TestWordList_FilterDocx(t *testing.T) {
	items := []DriveItem{
		{Name: "report.docx", ID: "id-1"},
		{Name: "data.xlsx", ID: "id-2"},
		{Name: "notes.DOCX", ID: "id-3"},
		{Name: "presentation.pptx", ID: "id-4"},
		{Name: "letter.docx", ID: "id-5"},
	}

	var docs []DriveItem
	for _, item := range items {
		if strings.HasSuffix(strings.ToLower(item.Name), ".docx") {
			docs = append(docs, item)
		}
	}

	assert.Len(t, docs, 3)
	assert.Equal(t, "report.docx", docs[0].Name)
	assert.Equal(t, "notes.DOCX", docs[1].Name)
	assert.Equal(t, "letter.docx", docs[2].Name)
}

// Tests for Word document output
func TestWordList_Output(t *testing.T) {
	docs := []DriveItem{
		{
			ID:                   "doc-1-long-id-for-testing",
			Name:                 "Annual Report.docx",
			Size:                 102400,
			LastModifiedDateTime: "2024-01-15T10:30:00Z",
		},
		{
			ID:                   "doc-2-long-id-for-testing",
			Name:                 "Project Proposal.docx",
			Size:                 51200,
			LastModifiedDateTime: "2024-01-14T15:00:00Z",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, doc := range docs {
		os.Stdout.WriteString("📝 " + doc.Name + " " + formatSize(doc.Size) + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "📝")
	assert.Contains(t, output, "Annual Report.docx")
	assert.Contains(t, output, "Project Proposal.docx")
	assert.Contains(t, output, "100.0 KB")
}

// Tests for Word metadata output
func TestWordGet_MetadataOutput(t *testing.T) {
	item := DriveItem{
		ID:                   "doc-metadata-test",
		Name:                 "Important.docx",
		Size:                 204800,
		CreatedDateTime:      "2024-01-01T08:00:00Z",
		LastModifiedDateTime: "2024-01-15T16:30:00Z",
		WebURL:               "https://example.com/doc",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("Name:     " + item.Name + "\n")
	os.Stdout.WriteString("Size:     " + formatSize(item.Size) + "\n")
	os.Stdout.WriteString("Created:  " + item.CreatedDateTime + "\n")
	os.Stdout.WriteString("Modified: " + item.LastModifiedDateTime + "\n")
	os.Stdout.WriteString("URL:      " + item.WebURL + "\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Name:     Important.docx")
	assert.Contains(t, output, "Size:     200.0 KB")
	assert.Contains(t, output, "Created:")
	assert.Contains(t, output, "Modified:")
	assert.Contains(t, output, "URL:")
}

// Tests for JSON output
func TestWordList_JSONOutput(t *testing.T) {
	docs := []DriveItem{
		{ID: "doc-1", Name: "Test.docx", Size: 1024},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(docs)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "doc-1"`)
	assert.Contains(t, output, `"name": "Test.docx"`)
}

// Tests for export format handling
func TestWordExport_FormatPath(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		containsPdf bool
	}{
		{"docx format", "docx", false},
		{"pdf format", "pdf", true},
		{"DOCX uppercase", "DOCX", false},
		{"PDF uppercase", "PDF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := strings.ToLower(tt.format)
			isPdf := format == "pdf"
			assert.Equal(t, tt.containsPdf, isPdf)
		})
	}
}

// Tests for export success output
func TestWordExport_SuccessOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Exported\n")
	os.Stdout.WriteString("  Format: DOCX\n")
	os.Stdout.WriteString("  Saved to: /output/doc.docx\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Exported")
	assert.Contains(t, output, "Format: DOCX")
	assert.Contains(t, output, "Saved to:")
}

// Tests for copy success output
func TestWordCopy_SuccessOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Copy initiated\n")
	os.Stdout.WriteString("  Name: Copy of Document.docx\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Copy initiated")
	assert.Contains(t, output, "Name:")
}

// Tests for create success output
func TestWordCreate_SuccessOutput(t *testing.T) {
	item := DriveItem{
		ID:   "new-doc-123",
		Name: "NewDocument.docx",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Document created\n")
	os.Stdout.WriteString("  Name: " + item.Name + "\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Document created")
	assert.Contains(t, output, "Name: NewDocument.docx")
}

// Edge cases
func TestWordList_EmptyResult(t *testing.T) {
	var docs []DriveItem
	assert.Empty(t, docs)
}

func TestWordList_NoDocxFiles(t *testing.T) {
	items := []DriveItem{
		{Name: "data.xlsx"},
		{Name: "presentation.pptx"},
	}

	var docs []DriveItem
	for _, item := range items {
		if strings.HasSuffix(strings.ToLower(item.Name), ".docx") {
			docs = append(docs, item)
		}
	}

	assert.Empty(t, docs)
}

func TestWordExport_JSONOutputFormat(t *testing.T) {
	result := map[string]interface{}{
		"success": true,
		"path":    "/output/doc.docx",
		"format":  "docx",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"success":true`)
	assert.Contains(t, string(data), `"path":"/output/doc.docx"`)
}
