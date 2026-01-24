package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Set test HOME
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	dir, err := ConfigDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "mog"), dir)
}

func TestConfig_SaveLoad(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config
	cfg := &Config{
		ClientID: "test-client-id-12345",
	}

	// Save
	err := Save(cfg)
	require.NoError(t, err)

	// Load
	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.ClientID, loaded.ClientID)
}

func TestConfig_LoadMissing(t *testing.T) {
	// Setup: use temp dir with no config
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load should return empty config, not error
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.ClientID)
}

func TestTokens_SaveLoad(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create tokens
	tokens := &Tokens{
		AccessToken:  "access-token-abc",
		RefreshToken: "refresh-token-xyz",
		ExpiresAt:    1234567890,
	}

	// Save
	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Load
	loaded, err := LoadTokens()
	require.NoError(t, err)
	assert.Equal(t, tokens.AccessToken, loaded.AccessToken)
	assert.Equal(t, tokens.RefreshToken, loaded.RefreshToken)
	assert.Equal(t, tokens.ExpiresAt, loaded.ExpiresAt)
}

func TestTokens_LoadMissing(t *testing.T) {
	// Setup: use temp dir with no tokens
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load should return error for missing tokens
	_, err := LoadTokens()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not logged in")
}

func TestTokens_Delete(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Save tokens first
	tokens := &Tokens{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresAt:    123,
	}
	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Delete
	err = DeleteTokens()
	require.NoError(t, err)

	// Should not exist anymore
	_, err = LoadTokens()
	assert.Error(t, err)
}

func TestTokens_DeleteMissing(t *testing.T) {
	// Setup: use temp dir with no tokens
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Delete should not error if file doesn't exist
	err := DeleteTokens()
	assert.NoError(t, err)
}

func TestSlugs_SaveLoad(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create slugs
	slugs := &Slugs{
		IDToSlug: map[string]string{
			"long-id-123": "abc123",
			"long-id-456": "def456",
		},
		SlugToID: map[string]string{
			"abc123": "long-id-123",
			"def456": "long-id-456",
		},
	}

	// Save
	err := SaveSlugs(slugs)
	require.NoError(t, err)

	// Load
	loaded, err := LoadSlugs()
	require.NoError(t, err)
	assert.Equal(t, slugs.IDToSlug, loaded.IDToSlug)
	assert.Equal(t, slugs.SlugToID, loaded.SlugToID)
}

func TestSlugs_LoadMissing(t *testing.T) {
	// Setup: use temp dir with no slugs
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load should return empty slugs, not error
	slugs, err := LoadSlugs()
	require.NoError(t, err)
	assert.NotNil(t, slugs)
	assert.NotNil(t, slugs.IDToSlug)
	assert.NotNil(t, slugs.SlugToID)
	assert.Empty(t, slugs.IDToSlug)
	assert.Empty(t, slugs.SlugToID)
}

func TestConfig_FilePermissions(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Save tokens (should have restricted permissions)
	tokens := &Tokens{AccessToken: "secret"}
	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Check file permissions
	configDir, _ := ConfigDir()
	info, err := os.Stat(filepath.Join(configDir, "tokens.json"))
	require.NoError(t, err)

	// Should be 0600 (owner read/write only)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestConfig_LoadCorruptedJSON(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir and corrupt file
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "settings.json"), []byte("{invalid json"), 0600))

	// Load should error
	_, err := Load()
	assert.Error(t, err)
}

func TestTokens_LoadCorruptedJSON(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir and corrupt file
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "tokens.json"), []byte("not json"), 0600))

	// Load should error
	_, err := LoadTokens()
	assert.Error(t, err)
}

func TestSlugs_LoadCorruptedJSON(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir and corrupt file
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "slugs.json"), []byte("{bad"), 0600))

	// Load should error
	_, err := LoadSlugs()
	assert.Error(t, err)
}

func TestSlugs_LoadPartialJSON(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir and partial file (missing maps)
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "slugs.json"), []byte("{}"), 0600))

	// Load should succeed and initialize maps
	slugs, err := LoadSlugs()
	require.NoError(t, err)
	assert.NotNil(t, slugs.IDToSlug)
	assert.NotNil(t, slugs.SlugToID)
}

func TestConfig_SaveCreatesDir(t *testing.T) {
	// Setup: use temp dir with no .config
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Save should create directory
	cfg := &Config{ClientID: "test"}
	err := Save(cfg)
	require.NoError(t, err)

	// Verify directory was created
	configDir := filepath.Join(tmpDir, ".config", "mog")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
