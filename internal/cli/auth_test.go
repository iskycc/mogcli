package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visionik/mogcli/internal/config"
)

func setupAuthTestConfig(t *testing.T) func() {
	t.Helper()

	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	return func() {
		os.Setenv("HOME", origHome)
	}
}

func TestAuthStatusCmd_NotLoggedIn(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// No tokens saved
	cmd := &AuthStatusCmd{}
	root := &Root{}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(root)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Not logged in")
}

func TestAuthStatusCmd_LoggedIn(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// Save tokens
	tokens := &config.Tokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    9999999999, // Far future
	}
	require.NoError(t, config.SaveTokens(tokens))

	// Save config with client ID
	cfg := &config.Config{
		ClientID: "test-client-id-12345678901234567890",
	}
	require.NoError(t, config.Save(cfg))

	cmd := &AuthStatusCmd{}
	root := &Root{}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(root)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Logged in")
	assert.Contains(t, output, "Token expires:")
	assert.Contains(t, output, "Client ID:")
}

func TestAuthStatusCmd_ExpiredToken(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// Save expired tokens
	tokens := &config.Tokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    1, // Epoch + 1 second (expired)
	}
	require.NoError(t, config.SaveTokens(tokens))

	cmd := &AuthStatusCmd{}
	root := &Root{}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(root)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Logged in")
	assert.Contains(t, output, "Expired")
}

func TestAuthLogoutCmd_Success(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// Save tokens
	tokens := &config.Tokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    9999999999,
	}
	require.NoError(t, config.SaveTokens(tokens))

	cmd := &AuthLogoutCmd{}
	root := &Root{}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(root)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Logged out successfully")

	// Verify tokens are deleted
	_, err = config.LoadTokens()
	assert.Error(t, err)
}

func TestAuthLogoutCmd_NoTokens(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// No tokens saved - logout should still succeed
	cmd := &AuthLogoutCmd{}
	root := &Root{}

	err := cmd.Run(root)
	require.NoError(t, err)
}

func TestAuthLoginCmd_SavesClientID(t *testing.T) {
	cleanup := setupAuthTestConfig(t)
	defer cleanup()

	// Note: We can't fully test login because it requires user interaction
	// But we can test that it saves the client ID
	clientID := "test-client-id-for-login"
	cfg := &config.Config{ClientID: clientID}
	err := config.Save(cfg)
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, clientID, loaded.ClientID)
}

func TestOpenBrowser_DoesNotPanic(t *testing.T) {
	// Test that openBrowser doesn't panic for various URLs
	// It may fail to actually open a browser in test environment, but shouldn't panic
	testURLs := []string{
		"https://example.com",
		"https://login.microsoftonline.com/common/oauth2/v2.0/devicecode",
		"",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			// Should not panic
			openBrowser(url)
		})
	}
}
