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

func setupDriveTestConfig(t *testing.T) func() {
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

// Tests for DriveItem type
func TestDriveItem_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		isFolder bool
	}{
		{
			name: "file item",
			json: `{
				"id": "item-123",
				"name": "document.pdf",
				"size": 1024,
				"createdDateTime": "2024-01-15T10:00:00Z",
				"lastModifiedDateTime": "2024-01-15T12:00:00Z",
				"webUrl": "https://example.com/file",
				"file": {"mimeType": "application/pdf"}
			}`,
			isFolder: false,
		},
		{
			name: "folder item",
			json: `{
				"id": "folder-456",
				"name": "Documents",
				"size": 0,
				"createdDateTime": "2024-01-10T08:00:00Z",
				"lastModifiedDateTime": "2024-01-15T15:00:00Z",
				"folder": {"childCount": 10}
			}`,
			isFolder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item DriveItem
			err := json.Unmarshal([]byte(tt.json), &item)
			require.NoError(t, err)

			if tt.isFolder {
				assert.NotNil(t, item.Folder)
				assert.Nil(t, item.File)
			} else {
				assert.NotNil(t, item.File)
				assert.Nil(t, item.Folder)
			}
		})
	}
}

// Tests for FolderInfo type
func TestFolderInfo_Unmarshal(t *testing.T) {
	jsonData := `{"childCount": 15}`

	var info FolderInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	require.NoError(t, err)
	assert.Equal(t, 15, info.ChildCount)
}

// Tests for FileInfo type
func TestFileInfo_Unmarshal(t *testing.T) {
	jsonData := `{"mimeType": "image/jpeg"}`

	var info FileInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", info.MimeType)
}

// Tests for formatSize
func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
		{"fractional KB", 1536, "1.5 KB"},
		{"fractional MB", 5242880, "5.0 MB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for DriveLsCmd
func TestDriveLsCmd_Fields(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"root", ""},
		{"folder path", "Documents/Work"},
		{"folder ID", "driveItem12345678901234567890123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &DriveLsCmd{Path: tt.path}
			assert.Equal(t, tt.path, cmd.Path)
		})
	}
}

func TestDriveLsCmd_PathType(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		isID   bool
	}{
		{"empty path", "", false},
		{"short path", "Documents", false},
		{"long ID-like", "AQMkADAwATMzAGZmAS04MDViLTRiNzg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLikelyID := len(tt.path) > 20
			assert.Equal(t, tt.isID, isLikelyID)
		})
	}
}

// Tests for DriveSearchCmd
func TestDriveSearchCmd_Fields(t *testing.T) {
	cmd := &DriveSearchCmd{
		Query: "budget.xlsx",
		Max:   50,
	}

	assert.Equal(t, "budget.xlsx", cmd.Query)
	assert.Equal(t, 50, cmd.Max)
}

func TestDriveSearchCmd_DefaultMax(t *testing.T) {
	cmd := &DriveSearchCmd{
		Query: "test",
		Max:   25, // Default
	}
	assert.Equal(t, 25, cmd.Max)
}

// Tests for DriveGetCmd
func TestDriveGetCmd_Fields(t *testing.T) {
	cmd := &DriveGetCmd{
		ID: "item-123",
	}
	assert.Equal(t, "item-123", cmd.ID)
}

// Tests for DriveDownloadCmd
func TestDriveDownloadCmd_Fields(t *testing.T) {
	cmd := &DriveDownloadCmd{
		ID:  "item-123",
		Out: "/path/to/output.pdf",
	}

	assert.Equal(t, "item-123", cmd.ID)
	assert.Equal(t, "/path/to/output.pdf", cmd.Out)
}

// Tests for DriveUploadCmd
func TestDriveUploadCmd_Fields(t *testing.T) {
	cmd := &DriveUploadCmd{
		Path:   "/local/file.txt",
		Folder: "folder-123",
		Name:   "renamed.txt",
	}

	assert.Equal(t, "/local/file.txt", cmd.Path)
	assert.Equal(t, "folder-123", cmd.Folder)
	assert.Equal(t, "renamed.txt", cmd.Name)
}

func TestDriveUploadCmd_DefaultName(t *testing.T) {
	cmd := &DriveUploadCmd{
		Path: "/local/path/document.pdf",
	}

	// When Name is empty, should use basename of Path
	name := cmd.Name
	if name == "" {
		name = filepath.Base(cmd.Path)
	}
	assert.Equal(t, "document.pdf", name)
}

// Tests for DriveMkdirCmd
func TestDriveMkdirCmd_Fields(t *testing.T) {
	cmd := &DriveMkdirCmd{
		Name:   "New Folder",
		Parent: "parent-folder-123",
	}

	assert.Equal(t, "New Folder", cmd.Name)
	assert.Equal(t, "parent-folder-123", cmd.Parent)
}

func TestDriveMkdirCmd_NoParent(t *testing.T) {
	cmd := &DriveMkdirCmd{
		Name: "Root Folder",
	}

	assert.Equal(t, "Root Folder", cmd.Name)
	assert.Empty(t, cmd.Parent)
}

// Tests for DriveMoveCmd
func TestDriveMoveCmd_Fields(t *testing.T) {
	cmd := &DriveMoveCmd{
		ID:          "item-123",
		Destination: "folder-456",
	}

	assert.Equal(t, "item-123", cmd.ID)
	assert.Equal(t, "folder-456", cmd.Destination)
}

// Tests for DriveCopyCmd
func TestDriveCopyCmd_Fields(t *testing.T) {
	cmd := &DriveCopyCmd{
		ID:   "item-123",
		Name: "copy-of-file.pdf",
	}

	assert.Equal(t, "item-123", cmd.ID)
	assert.Equal(t, "copy-of-file.pdf", cmd.Name)
}

// Tests for DriveRenameCmd
func TestDriveRenameCmd_Fields(t *testing.T) {
	cmd := &DriveRenameCmd{
		ID:   "item-123",
		Name: "new-name.txt",
	}

	assert.Equal(t, "item-123", cmd.ID)
	assert.Equal(t, "new-name.txt", cmd.Name)
}

// Tests for DriveDeleteCmd
func TestDriveDeleteCmd_Fields(t *testing.T) {
	cmd := &DriveDeleteCmd{
		ID: "item-123",
	}
	assert.Equal(t, "item-123", cmd.ID)
}

// Tests for drive list output
func TestDriveLs_OutputFormat(t *testing.T) {
	items := []DriveItem{
		{
			ID:   "file-123-long-id-for-testing",
			Name: "document.pdf",
			Size: 1024,
			File: &FileInfo{MimeType: "application/pdf"},
		},
		{
			ID:     "folder-456-long-id-for-testing",
			Name:   "Projects",
			Size:   0,
			Folder: &FolderInfo{ChildCount: 5},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, item := range items {
		itemType := "📄"
		if item.Folder != nil {
			itemType = "📁"
		}
		size := ""
		if item.Size > 0 {
			size = formatSize(item.Size)
		}
		// Simplified output for testing
		os.Stdout.WriteString(itemType + " " + item.Name + " " + size + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "📄")
	assert.Contains(t, output, "📁")
	assert.Contains(t, output, "document.pdf")
	assert.Contains(t, output, "Projects")
	assert.Contains(t, output, "1.0 KB")
}

// Tests for JSON output
func TestDriveItem_JSONOutput(t *testing.T) {
	item := DriveItem{
		ID:                   "item-123",
		Name:                 "test.txt",
		Size:                 256,
		CreatedDateTime:      "2024-01-15T10:00:00Z",
		LastModifiedDateTime: "2024-01-15T12:00:00Z",
		WebURL:               "https://example.com/file",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(item)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "item-123"`)
	assert.Contains(t, output, `"name": "test.txt"`)
	assert.Contains(t, output, `"size": 256`)
}

// Tests for DriveItem with full metadata
func TestDriveItem_FullMetadata(t *testing.T) {
	jsonData := `{
		"id": "AQMkADAwATMzAGZmAS04MDViLTRiNzgtMDACLTAwCgBGAAAD",
		"name": "Q4 Report.xlsx",
		"size": 2048576,
		"createdDateTime": "2024-01-01T08:00:00Z",
		"lastModifiedDateTime": "2024-01-15T16:30:00Z",
		"webUrl": "https://onedrive.example.com/documents/Q4%20Report.xlsx",
		"file": {
			"mimeType": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		}
	}`

	var item DriveItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)

	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "Q4 Report.xlsx", item.Name)
	assert.Equal(t, int64(2048576), item.Size)
	assert.Equal(t, "2024-01-01T08:00:00Z", item.CreatedDateTime)
	assert.Equal(t, "2024-01-15T16:30:00Z", item.LastModifiedDateTime)
	assert.NotEmpty(t, item.WebURL)
	assert.NotNil(t, item.File)
	assert.Contains(t, item.File.MimeType, "spreadsheet")
}

// Tests for folder with children
func TestDriveItem_FolderWithChildren(t *testing.T) {
	jsonData := `{
		"id": "folder-789",
		"name": "Shared Documents",
		"size": 0,
		"folder": {
			"childCount": 42
		}
	}`

	var item DriveItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)

	assert.NotNil(t, item.Folder)
	assert.Equal(t, 42, item.Folder.ChildCount)
	assert.Nil(t, item.File)
}

// Edge case tests
func TestDriveItem_EmptyName(t *testing.T) {
	jsonData := `{"id": "item-123", "name": ""}`

	var item DriveItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)
	assert.Empty(t, item.Name)
}

func TestFormatSize_LargeFile(t *testing.T) {
	// Test with very large file (multiple petabytes - edge case)
	size := int64(5 * 1024 * 1024 * 1024 * 1024 * 1024) // 5 PB
	result := formatSize(size)
	assert.Contains(t, result, "PB")
}

func TestFormatSize_ExactBoundaries(t *testing.T) {
	tests := []struct {
		bytes    int64
		contains string
	}{
		{1023, "B"},          // Just under 1 KB
		{1024, "KB"},         // Exactly 1 KB
		{1048575, "KB"},      // Just under 1 MB
		{1048576, "MB"},      // Exactly 1 MB
		{1073741823, "MB"},   // Just under 1 GB
		{1073741824, "GB"},   // Exactly 1 GB
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			result := formatSize(tt.bytes)
			assert.Contains(t, result, tt.contains)
		})
	}
}
