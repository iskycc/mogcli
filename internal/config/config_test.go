package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAccount(t *testing.T) {
	origAccount := currentAccount
	defer func() { currentAccount = origAccount }()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to default", "", "default"},
		{"named account", "work", "work"},
		{"personal account", "personal", "personal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetAccount(tt.input)
			assert.Equal(t, tt.expected, GetAccount())
		})
	}
}

func TestMigrateIfNeeded(t *testing.T) {
	t.Run("migrates legacy config", func(t *testing.T) {
		origHome := os.Getenv("HOME")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		// Create legacy files at base level
		baseDir := filepath.Join(tmpDir, ".config", "mog")
		require.NoError(t, os.MkdirAll(baseDir, 0700))
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, "tokens.json"), []byte(`{"access_token":"tok"}`), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, "settings.json"), []byte(`{"client_id":"cid"}`), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, "slugs.json"), []byte(`{}`), 0600))

		err := MigrateIfNeeded()
		require.NoError(t, err)

		// Legacy files should be moved to default/
		defaultDir := filepath.Join(baseDir, "default")
		_, err = os.Stat(filepath.Join(defaultDir, "tokens.json"))
		assert.NoError(t, err, "tokens.json should be in default/")
		_, err = os.Stat(filepath.Join(defaultDir, "settings.json"))
		assert.NoError(t, err, "settings.json should be in default/")
		_, err = os.Stat(filepath.Join(defaultDir, "slugs.json"))
		assert.NoError(t, err, "slugs.json should be in default/")

		// Legacy files should no longer exist at base
		_, err = os.Stat(filepath.Join(baseDir, "tokens.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("no-op when no legacy config", func(t *testing.T) {
		origHome := os.Getenv("HOME")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		err := MigrateIfNeeded()
		assert.NoError(t, err)
	})
}

func TestBaseConfigDir(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir, err := BaseConfigDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "mog"), dir)
}

func TestListAccounts(t *testing.T) {
	t.Run("lists accounts with tokens", func(t *testing.T) {
		origHome := os.Getenv("HOME")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		baseDir := filepath.Join(tmpDir, ".config", "mog")
		// Create two accounts with tokens
		for _, acct := range []string{"work", "personal"} {
			acctDir := filepath.Join(baseDir, acct)
			require.NoError(t, os.MkdirAll(acctDir, 0700))
			require.NoError(t, os.WriteFile(filepath.Join(acctDir, "tokens.json"), []byte(`{}`), 0600))
		}
		// Create a dir without tokens (should not be listed)
		require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "empty"), 0700))

		accounts, err := ListAccounts()
		require.NoError(t, err)
		assert.Len(t, accounts, 2)
		assert.Contains(t, accounts, "work")
		assert.Contains(t, accounts, "personal")
	})

	t.Run("empty when no config dir", func(t *testing.T) {
		origHome := os.Getenv("HOME")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		accounts, err := ListAccounts()
		require.NoError(t, err)
		assert.Empty(t, accounts)
	})
}

func TestConfigDir(t *testing.T) {
	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Set test HOME
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	dir, err := ConfigDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "mog", currentAccount), dir)
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
		Region:   "china",
	}

	// Save
	err := Save(cfg)
	require.NoError(t, err)

	// Load
	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.ClientID, loaded.ClientID)
	assert.Equal(t, "china", loaded.Region)
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
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
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
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
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
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
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

func TestConfig_GetClientID(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name:     "Go format (client_id)",
			cfg:      &Config{ClientID: "go-client-id"},
			expected: "go-client-id",
		},
		{
			name:     "Node format (clientId)",
			cfg:      &Config{ClientIDv2: "node-client-id"},
			expected: "node-client-id",
		},
		{
			name:     "Both formats - Go takes precedence",
			cfg:      &Config{ClientID: "go-id", ClientIDv2: "node-id"},
			expected: "go-id",
		},
		{
			name:     "Empty config",
			cfg:      &Config{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetClientID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetRegion(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name:     "china region",
			cfg:      &Config{Region: "china"},
			expected: "china",
		},
		{
			name:     "global region",
			cfg:      &Config{Region: "global"},
			expected: "global",
		},
		{
			name:     "empty defaults to global",
			cfg:      &Config{},
			expected: "global",
		},
		{
			name:     "legacy config without region field",
			cfg:      &Config{ClientID: "legacy-client-id"},
			expected: "global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetRegion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokens_GetExpiresAt(t *testing.T) {
	tests := []struct {
		name     string
		tokens   *Tokens
		expected int64
	}{
		{
			name:     "Go format (expires_at)",
			tokens:   &Tokens{ExpiresAt: 1234567890},
			expected: 1234567890,
		},
		{
			name: "Node format (saved_at + expires_in)",
			tokens: &Tokens{
				SavedAt:   1234567890000, // ms
				ExpiresIn: 3600,          // seconds
			},
			expected: 1234567890 + 3600,
		},
		{
			name:     "Both formats - Go takes precedence",
			tokens:   &Tokens{ExpiresAt: 9999, SavedAt: 1000000, ExpiresIn: 100},
			expected: 9999,
		},
		{
			name:     "Empty tokens",
			tokens:   &Tokens{},
			expected: 0,
		},
		{
			name:     "Node format - missing saved_at",
			tokens:   &Tokens{ExpiresIn: 3600},
			expected: 0,
		},
		{
			name:     "Node format - missing expires_in",
			tokens:   &Tokens{SavedAt: 1234567890000},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tokens.GetExpiresAt()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokens_SaveCreatesDir(t *testing.T) {
	// Setup: use temp dir with no .config
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Save should create directory
	tokens := &Tokens{AccessToken: "test"}
	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Verify directory was created
	configDir := filepath.Join(tmpDir, ".config", "mog")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSlugs_SaveCreatesDir(t *testing.T) {
	// Setup: use temp dir with no .config
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Save should create directory
	slugs := &Slugs{
		IDToSlug: map[string]string{"id": "slug"},
		SlugToID: map[string]string{"slug": "id"},
	}
	err := SaveSlugs(slugs)
	require.NoError(t, err)

	// Verify directory was created
	configDir := filepath.Join(tmpDir, ".config", "mog")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestConfigDir_Success(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	dir, err := ConfigDir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".config/mog")
}

func TestLoadTokens_ReadError(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with a directory named tokens.json (can't read a directory as file)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "tokens.json"), 0700))

	_, err := LoadTokens()
	assert.Error(t, err)
}

func TestLoad_ReadError(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with a directory named settings.json (can't read a directory as file)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "settings.json"), 0700))

	_, err := Load()
	assert.Error(t, err)
}

func TestLoadSlugs_ReadError(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with a directory named slugs.json (can't read a directory as file)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "slugs.json"), 0700))

	_, err := LoadSlugs()
	assert.Error(t, err)
}

func TestDeleteTokens_Error(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with a directory named tokens.json (can't delete a directory with Remove)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "tokens.json"), 0700))
	// Put something in the directory so Remove fails
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "tokens.json", "dummy"), []byte("x"), 0600))

	err := DeleteTokens()
	assert.Error(t, err)
}

func TestSlugs_OnlyIDToSlug(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with partial slugs (only id_to_slug)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "slugs.json"), []byte(`{"id_to_slug": {"id1": "slug1"}}`), 0600))

	slugs, err := LoadSlugs()
	require.NoError(t, err)
	assert.NotNil(t, slugs.IDToSlug)
	assert.NotNil(t, slugs.SlugToID)
	assert.Equal(t, "slug1", slugs.IDToSlug["id1"])
}

func TestAliases_SaveLoad(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	aliases := &Aliases{
		NameToTarget: map[string]string{
			"standup": "a3f2c891",
			"budget":  "AQMkADAwATMzAGZmAS04MDViLTRiNzgtMDA...",
		},
	}

	err := SaveAliases(aliases)
	require.NoError(t, err)

	loaded, err := LoadAliases()
	require.NoError(t, err)
	assert.Equal(t, aliases.NameToTarget, loaded.NameToTarget)
}

func TestAliases_LoadMissing(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	aliases, err := LoadAliases()
	require.NoError(t, err)
	assert.NotNil(t, aliases)
	assert.NotNil(t, aliases.NameToTarget)
	assert.Empty(t, aliases.NameToTarget)
}

func TestAliases_LoadCorruptedJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "aliases.json"), []byte("{bad"), 0600))

	_, err := LoadAliases()
	assert.Error(t, err)
}

func TestAliases_LoadPartialJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "aliases.json"), []byte("{}"), 0600))

	aliases, err := LoadAliases()
	require.NoError(t, err)
	assert.NotNil(t, aliases.NameToTarget)
}

func TestAliases_SaveCreatesDir(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	aliases := &Aliases{
		NameToTarget: map[string]string{"test": "value"},
	}
	err := SaveAliases(aliases)
	require.NoError(t, err)

	configDir := filepath.Join(tmpDir, ".config", "mog")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestAliases_ReadError(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "aliases.json"), 0700))

	_, err := LoadAliases()
	assert.Error(t, err)
}

func TestSlugs_OnlySlugToID(t *testing.T) {
	// Setup: use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir with partial slugs (only slug_to_id)
	configDir := filepath.Join(tmpDir, ".config", "mog", currentAccount)
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "slugs.json"), []byte(`{"slug_to_id": {"slug1": "id1"}}`), 0600))

	slugs, err := LoadSlugs()
	require.NoError(t, err)
	assert.NotNil(t, slugs.IDToSlug)
	assert.NotNil(t, slugs.SlugToID)
	assert.Equal(t, "id1", slugs.SlugToID["slug1"])
}
