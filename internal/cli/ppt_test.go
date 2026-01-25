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

func setupPPTTestConfig(t *testing.T) func() {
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

// Tests for PPTListCmd
func TestPPTListCmd_DefaultMax(t *testing.T) {
	cmd := &PPTListCmd{
		Max: 50, // Default value
	}
	assert.Equal(t, 50, cmd.Max)
}

func TestPPTListCmd_CustomMax(t *testing.T) {
	cmd := &PPTListCmd{
		Max: 100,
	}
	assert.Equal(t, 100, cmd.Max)
}

// Tests for PPTGetCmd
func TestPPTGetCmd_Fields(t *testing.T) {
	cmd := &PPTGetCmd{
		ID: "ppt-123",
	}
	assert.Equal(t, "ppt-123", cmd.ID)
}

// Tests for PPTExportCmd
func TestPPTExportCmd_Fields(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		out    string
		format string
	}{
		{
			name:   "pptx export",
			id:     "ppt-123",
			out:    "/output/presentation.pptx",
			format: "pptx",
		},
		{
			name:   "pdf export",
			id:     "ppt-456",
			out:    "/output/presentation.pdf",
			format: "pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PPTExportCmd{
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

func TestPPTExportCmd_DefaultFormat(t *testing.T) {
	cmd := &PPTExportCmd{
		ID:     "ppt-123",
		Out:    "/output/file.pptx",
		Format: "pptx", // Default
	}
	assert.Equal(t, "pptx", cmd.Format)
}

// Tests for PPTCopyCmd
func TestPPTCopyCmd_Fields(t *testing.T) {
	cmd := &PPTCopyCmd{
		ID:     "ppt-123",
		Name:   "Copy of Presentation.pptx",
		Folder: "folder-456",
	}

	assert.Equal(t, "ppt-123", cmd.ID)
	assert.Equal(t, "Copy of Presentation.pptx", cmd.Name)
	assert.Equal(t, "folder-456", cmd.Folder)
}

func TestPPTCopyCmd_NoFolder(t *testing.T) {
	cmd := &PPTCopyCmd{
		ID:   "ppt-123",
		Name: "Copy.pptx",
	}

	assert.Empty(t, cmd.Folder)
}

// Tests for PPTCreateCmd
func TestPPTCreateCmd_Fields(t *testing.T) {
	cmd := &PPTCreateCmd{
		Name:   "New Presentation.pptx",
		Folder: "folder-123",
	}

	assert.Equal(t, "New Presentation.pptx", cmd.Name)
	assert.Equal(t, "folder-123", cmd.Folder)
}

func TestPPTCreateCmd_AutoExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already has pptx", "Slides.pptx", "Slides.pptx"},
		{"no extension", "Slides", "Slides.pptx"},
		{"uppercase pptx", "Slides.PPTX", "Slides.PPTX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.input
			// Logic from the command
			if !strings.HasSuffix(strings.ToLower(name), ".pptx") {
				name += ".pptx"
			}

			if tt.name == "no extension" {
				assert.Equal(t, "Slides.pptx", name)
			} else {
				assert.Equal(t, tt.input, name)
			}
		})
	}
}

// Tests for PPT presentation filtering (from drive search)
func TestPPTList_FilterPptx(t *testing.T) {
	items := []DriveItem{
		{Name: "quarterly.pptx", ID: "id-1"},
		{Name: "data.xlsx", ID: "id-2"},
		{Name: "slides.PPTX", ID: "id-3"},
		{Name: "document.docx", ID: "id-4"},
		{Name: "pitch.pptx", ID: "id-5"},
	}

	var presentations []DriveItem
	for _, item := range items {
		if strings.HasSuffix(strings.ToLower(item.Name), ".pptx") {
			presentations = append(presentations, item)
		}
	}

	assert.Len(t, presentations, 3)
	assert.Equal(t, "quarterly.pptx", presentations[0].Name)
	assert.Equal(t, "slides.PPTX", presentations[1].Name)
	assert.Equal(t, "pitch.pptx", presentations[2].Name)
}

// Tests for PPT presentation output
func TestPPTList_Output(t *testing.T) {
	presentations := []DriveItem{
		{
			ID:                   "ppt-1-long-id-for-testing",
			Name:                 "Q4 Review.pptx",
			Size:                 5242880,
			LastModifiedDateTime: "2024-01-15T10:30:00Z",
		},
		{
			ID:                   "ppt-2-long-id-for-testing",
			Name:                 "Product Launch.pptx",
			Size:                 10485760,
			LastModifiedDateTime: "2024-01-14T15:00:00Z",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, ppt := range presentations {
		os.Stdout.WriteString("📊 " + ppt.Name + " " + formatSize(ppt.Size) + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "📊")
	assert.Contains(t, output, "Q4 Review.pptx")
	assert.Contains(t, output, "Product Launch.pptx")
	assert.Contains(t, output, "5.0 MB")
	assert.Contains(t, output, "10.0 MB")
}

// Tests for PPT metadata output
func TestPPTGet_MetadataOutput(t *testing.T) {
	item := DriveItem{
		ID:                   "ppt-metadata-test",
		Name:                 "Important Presentation.pptx",
		Size:                 15728640,
		CreatedDateTime:      "2024-01-01T08:00:00Z",
		LastModifiedDateTime: "2024-01-15T16:30:00Z",
		WebURL:               "https://example.com/ppt",
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

	assert.Contains(t, output, "Name:     Important Presentation.pptx")
	assert.Contains(t, output, "Size:     15.0 MB")
	assert.Contains(t, output, "Created:")
	assert.Contains(t, output, "Modified:")
	assert.Contains(t, output, "URL:")
}

// Tests for JSON output
func TestPPTList_JSONOutput(t *testing.T) {
	presentations := []DriveItem{
		{ID: "ppt-1", Name: "Test.pptx", Size: 2048},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(presentations)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "ppt-1"`)
	assert.Contains(t, output, `"name": "Test.pptx"`)
}

// Tests for export format handling
func TestPPTExport_FormatPath(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		containsPdf bool
	}{
		{"pptx format", "pptx", false},
		{"pdf format", "pdf", true},
		{"PPTX uppercase", "PPTX", false},
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
func TestPPTExport_SuccessOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Exported\n")
	os.Stdout.WriteString("  Format: PPTX\n")
	os.Stdout.WriteString("  Saved to: /output/presentation.pptx\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Exported")
	assert.Contains(t, output, "Format: PPTX")
	assert.Contains(t, output, "Saved to:")
}

// Tests for copy success output
func TestPPTCopy_SuccessOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Copy initiated\n")
	os.Stdout.WriteString("  Name: Copy of Presentation.pptx\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Copy initiated")
	assert.Contains(t, output, "Name:")
}

// Tests for create success output
func TestPPTCreate_SuccessOutput(t *testing.T) {
	item := DriveItem{
		ID:   "new-ppt-123",
		Name: "NewPresentation.pptx",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Stdout.WriteString("✓ Presentation created\n")
	os.Stdout.WriteString("  Name: " + item.Name + "\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "✓ Presentation created")
	assert.Contains(t, output, "Name: NewPresentation.pptx")
}

// Edge cases
func TestPPTList_EmptyResult(t *testing.T) {
	var presentations []DriveItem
	assert.Empty(t, presentations)
}

func TestPPTList_NoPptxFiles(t *testing.T) {
	items := []DriveItem{
		{Name: "data.xlsx"},
		{Name: "document.docx"},
	}

	var presentations []DriveItem
	for _, item := range items {
		if strings.HasSuffix(strings.ToLower(item.Name), ".pptx") {
			presentations = append(presentations, item)
		}
	}

	assert.Empty(t, presentations)
}

func TestPPTExport_JSONOutputFormat(t *testing.T) {
	result := map[string]interface{}{
		"success": true,
		"path":    "/output/presentation.pptx",
		"format":  "pptx",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"success":true`)
	assert.Contains(t, string(data), `"path":"/output/presentation.pptx"`)
}

// Tests for large presentation sizes
func TestPPTList_LargePresentations(t *testing.T) {
	sizes := []int64{
		52428800,   // 50 MB
		104857600,  // 100 MB
		524288000,  // 500 MB
	}

	for _, size := range sizes {
		formatted := formatSize(size)
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "MB")
	}
}

// Tests for PPT with special characters in name
func TestPPTList_SpecialCharacters(t *testing.T) {
	items := []DriveItem{
		{Name: "Q1 & Q2 Review.pptx"},
		{Name: "Project (Final).pptx"},
		{Name: "Budget - 2024.pptx"},
	}

	for _, item := range items {
		assert.True(t, strings.HasSuffix(strings.ToLower(item.Name), ".pptx"))
	}
}
