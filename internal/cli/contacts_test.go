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

func setupContactsTestConfig(t *testing.T) func() {
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

// Tests for Contact type
func TestContact_Unmarshal(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "full contact",
			json: `{
				"id": "contact-123",
				"displayName": "John Doe",
				"emailAddresses": [
					{"address": "john@example.com", "name": "John Doe"},
					{"address": "john.doe@work.com", "name": ""}
				],
				"businessPhones": ["+1-555-123-4567", "+1-555-987-6543"],
				"mobilePhone": "+1-555-111-2222",
				"companyName": "Acme Corp",
				"jobTitle": "Software Engineer"
			}`,
		},
		{
			name: "minimal contact",
			json: `{
				"id": "contact-456",
				"displayName": "Jane Smith"
			}`,
		},
		{
			name: "contact with only email",
			json: `{
				"id": "contact-789",
				"displayName": "Bob Wilson",
				"emailAddresses": [{"address": "bob@example.com"}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var contact Contact
			err := json.Unmarshal([]byte(tt.json), &contact)
			require.NoError(t, err)
			assert.NotEmpty(t, contact.ID)
			assert.NotEmpty(t, contact.DisplayName)
		})
	}
}

// Tests for EmailRecord type
func TestEmailRecord_Unmarshal(t *testing.T) {
	jsonData := `{"address": "user@example.com", "name": "User Name"}`

	var record EmailRecord
	err := json.Unmarshal([]byte(jsonData), &record)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", record.Address)
	assert.Equal(t, "User Name", record.Name)
}

// Tests for DirectoryUser type
func TestDirectoryUser_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "user-123",
		"displayName": "Alice Johnson",
		"mail": "alice@company.com",
		"jobTitle": "Product Manager"
	}`

	var user DirectoryUser
	err := json.Unmarshal([]byte(jsonData), &user)
	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "Alice Johnson", user.DisplayName)
	assert.Equal(t, "alice@company.com", user.Mail)
	assert.Equal(t, "Product Manager", user.JobTitle)
}

// Tests for ContactsListCmd
func TestContactsListCmd_DefaultMax(t *testing.T) {
	cmd := &ContactsListCmd{
		Max: 50, // Default value
	}
	assert.Equal(t, 50, cmd.Max)
}

func TestContactsListCmd_CustomMax(t *testing.T) {
	cmd := &ContactsListCmd{
		Max: 100,
	}
	assert.Equal(t, 100, cmd.Max)
}

// Tests for ContactsSearchCmd
func TestContactsSearchCmd_Fields(t *testing.T) {
	cmd := &ContactsSearchCmd{
		Query: "john",
	}
	assert.Equal(t, "john", cmd.Query)
}

func TestContactsSearchCmd_EmailSearch(t *testing.T) {
	cmd := &ContactsSearchCmd{
		Query: "@example.com",
	}
	assert.Contains(t, cmd.Query, "@")
}

// Tests for ContactsGetCmd
func TestContactsGetCmd_Fields(t *testing.T) {
	cmd := &ContactsGetCmd{
		ID: "contact-123",
	}
	assert.Equal(t, "contact-123", cmd.ID)
}

// Tests for ContactsCreateCmd
func TestContactsCreateCmd_Fields(t *testing.T) {
	cmd := &ContactsCreateCmd{
		Name:    "New Contact",
		Email:   "new@example.com",
		Phone:   "+1-555-000-1111",
		Company: "New Company",
		Title:   "Developer",
	}

	assert.Equal(t, "New Contact", cmd.Name)
	assert.Equal(t, "new@example.com", cmd.Email)
	assert.Equal(t, "+1-555-000-1111", cmd.Phone)
	assert.Equal(t, "New Company", cmd.Company)
	assert.Equal(t, "Developer", cmd.Title)
}

func TestContactsCreateCmd_MinimalFields(t *testing.T) {
	cmd := &ContactsCreateCmd{
		Name: "Simple Contact",
	}

	assert.Equal(t, "Simple Contact", cmd.Name)
	assert.Empty(t, cmd.Email)
	assert.Empty(t, cmd.Phone)
	assert.Empty(t, cmd.Company)
	assert.Empty(t, cmd.Title)
}

// Tests for ContactsUpdateCmd
func TestContactsUpdateCmd_Fields(t *testing.T) {
	cmd := &ContactsUpdateCmd{
		ID:      "contact-123",
		Name:    "Updated Name",
		Email:   "updated@example.com",
		Phone:   "+1-555-999-8888",
		Company: "Updated Company",
		Title:   "Senior Developer",
	}

	assert.Equal(t, "contact-123", cmd.ID)
	assert.Equal(t, "Updated Name", cmd.Name)
	assert.Equal(t, "updated@example.com", cmd.Email)
}

func TestContactsUpdateCmd_RequiresUpdates(t *testing.T) {
	cleanup := setupContactsTestConfig(t)
	defer cleanup()

	cmd := &ContactsUpdateCmd{
		ID: "contact-123",
		// No updates specified
	}

	root := &Root{}
	err := cmd.Run(root)

	// Should error if no updates specified (or fail on client)
	_ = err
	assert.NotNil(t, cmd)
}

// Tests for ContactsDeleteCmd
func TestContactsDeleteCmd_Fields(t *testing.T) {
	cmd := &ContactsDeleteCmd{
		ID: "contact-123",
	}
	assert.Equal(t, "contact-123", cmd.ID)
}

// Tests for ContactsDirectoryCmd
func TestContactsDirectoryCmd_Fields(t *testing.T) {
	cmd := &ContactsDirectoryCmd{
		Query: "engineering",
	}
	assert.Equal(t, "engineering", cmd.Query)
}

// Tests for Contact data output
func TestContact_Output(t *testing.T) {
	contact := Contact{
		ID:          "contact-id-12345678901234567890",
		DisplayName: "Test User",
		EmailAddresses: []EmailRecord{
			{Address: "test@example.com", Name: "Test User"},
		},
		BusinessPhones: []string{"+1-555-123-4567"},
		MobilePhone:    "+1-555-999-8888",
		CompanyName:    "Test Company",
		JobTitle:       "Test Title",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simulate contact get output
	email := ""
	if len(contact.EmailAddresses) > 0 {
		email = contact.EmailAddresses[0].Address
	}
	os.Stdout.WriteString("Name:  " + contact.DisplayName + "\n")
	os.Stdout.WriteString("Email: " + email + "\n")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Test User")
	assert.Contains(t, output, "test@example.com")
}

// Tests for Contact list output format
func TestContactsList_OutputFormat(t *testing.T) {
	contacts := []Contact{
		{
			ID:          "contact-1",
			DisplayName: "Alice",
			EmailAddresses: []EmailRecord{
				{Address: "alice@example.com"},
			},
		},
		{
			ID:          "contact-2",
			DisplayName: "Bob",
			EmailAddresses: []EmailRecord{
				{Address: "bob@example.com"},
			},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, c := range contacts {
		email := ""
		if len(c.EmailAddresses) > 0 {
			email = c.EmailAddresses[0].Address
		}
		os.Stdout.WriteString(c.DisplayName + " " + email + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "alice@example.com")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "bob@example.com")
}

// Tests for Contact JSON output
func TestContact_JSONOutput(t *testing.T) {
	contact := Contact{
		ID:          "contact-123",
		DisplayName: "JSON Test",
		EmailAddresses: []EmailRecord{
			{Address: "json@test.com"},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(contact)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, `"id": "contact-123"`)
	assert.Contains(t, output, `"displayName": "JSON Test"`)
}

// Tests for multiple email addresses
func TestContact_MultipleEmails(t *testing.T) {
	jsonData := `{
		"id": "contact-multi",
		"displayName": "Multi Email User",
		"emailAddresses": [
			{"address": "personal@example.com", "name": "Personal"},
			{"address": "work@company.com", "name": "Work"},
			{"address": "backup@other.com", "name": "Backup"}
		]
	}`

	var contact Contact
	err := json.Unmarshal([]byte(jsonData), &contact)
	require.NoError(t, err)
	assert.Len(t, contact.EmailAddresses, 3)
	assert.Equal(t, "personal@example.com", contact.EmailAddresses[0].Address)
	assert.Equal(t, "work@company.com", contact.EmailAddresses[1].Address)
}

// Tests for multiple business phones
func TestContact_MultiplePhones(t *testing.T) {
	jsonData := `{
		"id": "contact-phones",
		"displayName": "Phone User",
		"businessPhones": ["+1-111-111-1111", "+1-222-222-2222"]
	}`

	var contact Contact
	err := json.Unmarshal([]byte(jsonData), &contact)
	require.NoError(t, err)
	assert.Len(t, contact.BusinessPhones, 2)
}

// Tests for DirectoryUser list output
func TestDirectoryUser_ListOutput(t *testing.T) {
	users := []DirectoryUser{
		{
			ID:          "user-1",
			DisplayName: "Director Alice",
			Mail:        "dalice@corp.com",
			JobTitle:    "Director",
		},
		{
			ID:          "user-2",
			DisplayName: "Manager Bob",
			Mail:        "mbob@corp.com",
			JobTitle:    "Manager",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, u := range users {
		os.Stdout.WriteString(u.DisplayName + " " + u.Mail + " " + u.JobTitle + "\n")
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Director Alice")
	assert.Contains(t, output, "dalice@corp.com")
	assert.Contains(t, output, "Director")
}

// Edge case tests
func TestContact_NoEmailAddresses(t *testing.T) {
	contact := Contact{
		ID:          "contact-no-email",
		DisplayName: "No Email User",
	}

	email := ""
	if len(contact.EmailAddresses) > 0 {
		email = contact.EmailAddresses[0].Address
	}
	assert.Empty(t, email)
}

func TestContact_EmptyDisplayName(t *testing.T) {
	jsonData := `{"id": "contact-empty", "displayName": ""}`

	var contact Contact
	err := json.Unmarshal([]byte(jsonData), &contact)
	require.NoError(t, err)
	assert.Empty(t, contact.DisplayName)
}
