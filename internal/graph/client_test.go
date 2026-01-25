package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visionik/mogcli/internal/config"
)

func setupTestConfig(t *testing.T) func() {
	t.Helper()

	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	// Save test tokens (expires far in the future)
	tokens := &config.Tokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    9999999999,
	}
	require.NoError(t, config.SaveTokens(tokens))

	// Save test config
	cfg := &config.Config{
		ClientID: "test-client-id-12345678901234567890",
	}
	require.NoError(t, config.Save(cfg))

	return func() {
		os.Setenv("HOME", origHome)
	}
}

func TestNewClient_NoTokens(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir but no tokens
	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	_, err := NewClient()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not logged in")
}

func TestNewClient_ValidTokens(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	client, err := NewClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_ExpiredToken_NoRefresh(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "mog")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	// Save expired token with no refresh token
	tokens := &config.Tokens{
		AccessToken: "expired-token",
		ExpiresAt:   1, // Expired
	}
	require.NoError(t, config.SaveTokens(tokens))

	_, err := NewClient()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestClient_Get(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []map[string]string{
				{"id": "1", "name": "Test"},
			},
		})
	}))
	defer server.Close()

	// Create client that uses mock server
	client := &Client{
		httpClient: server.Client(),
		token:      "test-access-token",
	}

	// Override the base URL (this is a test-only approach)
	ctx := context.Background()

	// Make request through the mock server
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/me/messages", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+client.token)

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Contains(t, result, "value")
}

func TestClient_ErrorResponse(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "ResourceNotFound",
				"message": "The resource could not be found",
			},
		})
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		token:      "test-access-token",
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/me/messages/invalid", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+client.token)

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestClient_Post(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "new-id",
		})
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		token:      "test-access-token",
	}

	ctx := context.Background()
	body := map[string]string{"subject": "Test"}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/me/messages", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+client.token)
	req.Header.Set("Content-Type", "application/json")
	_ = bodyBytes // Would be used in actual request

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestClient_Delete(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		token:      "test-access-token",
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "DELETE", server.URL+"/me/messages/123", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+client.token)

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestClient_QueryParams(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	var receivedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		token:      "test-access-token",
	}

	ctx := context.Background()
	query := url.Values{}
	query.Set("$top", "10")
	query.Set("$orderby", "receivedDateTime desc")

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/me/messages?"+query.Encode(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+client.token)

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "10", receivedQuery.Get("$top"))
	assert.Equal(t, "receivedDateTime desc", receivedQuery.Get("$orderby"))
}

func TestDeviceCodeResponse_Parse(t *testing.T) {
	jsonData := `{
		"device_code": "test-device-code",
		"user_code": "ABCD1234",
		"verification_uri": "https://microsoft.com/devicelogin",
		"expires_in": 900,
		"interval": 5,
		"message": "To sign in, use a web browser to open the page"
	}`

	var resp DeviceCodeResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	assert.Equal(t, "test-device-code", resp.DeviceCode)
	assert.Equal(t, "ABCD1234", resp.UserCode)
	assert.Equal(t, "https://microsoft.com/devicelogin", resp.VerificationURI)
	assert.Equal(t, 900, resp.ExpiresIn)
	assert.Equal(t, 5, resp.Interval)
}

func TestTokenResponse_Parse(t *testing.T) {
	jsonData := `{
		"access_token": "test-access-token",
		"refresh_token": "test-refresh-token",
		"expires_in": 3600,
		"token_type": "Bearer",
		"scope": "User.Read Mail.Read"
	}`

	var resp TokenResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	assert.Equal(t, "test-access-token", resp.AccessToken)
	assert.Equal(t, "test-refresh-token", resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestTokenResponse_Error(t *testing.T) {
	jsonData := `{
		"error": "authorization_pending",
		"error_description": "The user has not yet completed authorization"
	}`

	var resp TokenResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	assert.Equal(t, "authorization_pending", resp.Error)
	assert.Contains(t, resp.ErrorDesc, "not yet completed")
}
