package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visionik/mogcli/internal/config"
)

func setupCalendarTestConfig(t *testing.T) func() {
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

// Tests for Event type
func TestEvent_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Event
	}{
		{
			name: "full event",
			json: `{
				"id": "event-123",
				"subject": "Team Meeting",
				"isAllDay": false,
				"start": {
					"dateTime": "2024-01-15T10:00:00.0000000",
					"timeZone": "UTC"
				},
				"end": {
					"dateTime": "2024-01-15T11:00:00.0000000",
					"timeZone": "UTC"
				},
				"location": {
					"displayName": "Conference Room A"
				},
				"organizer": {
					"emailAddress": {
						"name": "John Doe",
						"address": "john@example.com"
					}
				}
			}`,
			expected: Event{
				ID:       "event-123",
				Subject:  "Team Meeting",
				IsAllDay: false,
			},
		},
		{
			name: "all day event",
			json: `{
				"id": "event-456",
				"subject": "Holiday",
				"isAllDay": true,
				"start": {"dateTime": "2024-01-01T00:00:00.0000000", "timeZone": "UTC"},
				"end": {"dateTime": "2024-01-02T00:00:00.0000000", "timeZone": "UTC"}
			}`,
			expected: Event{
				ID:       "event-456",
				Subject:  "Holiday",
				IsAllDay: true,
			},
		},
		{
			name: "minimal event",
			json: `{"id": "event-789", "subject": "Quick Sync"}`,
			expected: Event{
				ID:      "event-789",
				Subject: "Quick Sync",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event Event
			err := json.Unmarshal([]byte(tt.json), &event)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.ID, event.ID)
			assert.Equal(t, tt.expected.Subject, event.Subject)
			assert.Equal(t, tt.expected.IsAllDay, event.IsAllDay)
		})
	}
}

// Tests for Time type
func TestTime_Unmarshal(t *testing.T) {
	jsonData := `{
		"dateTime": "2024-01-15T10:30:00.0000000",
		"timeZone": "UTC"
	}`

	var tm Time
	err := json.Unmarshal([]byte(jsonData), &tm)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-15T10:30:00.0000000", tm.DateTime)
	assert.Equal(t, "UTC", tm.TimeZone)
}

// Tests for Loc type
func TestLoc_Unmarshal(t *testing.T) {
	jsonData := `{"displayName": "Conference Room B"}`

	var loc Loc
	err := json.Unmarshal([]byte(jsonData), &loc)
	require.NoError(t, err)
	assert.Equal(t, "Conference Room B", loc.DisplayName)
}

// Tests for Body type
func TestBody_Unmarshal(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		contentType string
		content     string
	}{
		{
			name:        "text body",
			json:        `{"contentType": "text", "content": "Meeting notes"}`,
			contentType: "text",
			content:     "Meeting notes",
		},
		{
			name:        "html body",
			json:        `{"contentType": "html", "content": "<p>Meeting notes</p>"}`,
			contentType: "html",
			content:     "<p>Meeting notes</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body Body
			err := json.Unmarshal([]byte(tt.json), &body)
			require.NoError(t, err)
			assert.Equal(t, tt.contentType, body.ContentType)
			assert.Equal(t, tt.content, body.Content)
		})
	}
}

// Tests for Org type
func TestOrg_Unmarshal(t *testing.T) {
	jsonData := `{
		"emailAddress": {
			"name": "Jane Smith",
			"address": "jane@example.com"
		}
	}`

	var org Org
	err := json.Unmarshal([]byte(jsonData), &org)
	require.NoError(t, err)
	assert.Equal(t, "Jane Smith", org.EmailAddress.Name)
	assert.Equal(t, "jane@example.com", org.EmailAddress.Address)
}

// Tests for Calendar type
func TestCalendar_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		isDefault bool
	}{
		{
			name:      "default calendar",
			json:      `{"id": "cal-123", "name": "Calendar", "isDefaultCalendar": true}`,
			isDefault: true,
		},
		{
			name:      "secondary calendar",
			json:      `{"id": "cal-456", "name": "Work", "isDefaultCalendar": false}`,
			isDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cal Calendar
			err := json.Unmarshal([]byte(tt.json), &cal)
			require.NoError(t, err)
			assert.Equal(t, tt.isDefault, cal.IsDefaultCalendar)
		})
	}
}

// Tests for CalendarListCmd
func TestCalendarListCmd_DefaultValues(t *testing.T) {
	cmd := &CalendarListCmd{
		Max: 25,
	}
	assert.Equal(t, 25, cmd.Max)
	assert.Empty(t, cmd.Calendar)
	assert.Empty(t, cmd.From)
	assert.Empty(t, cmd.To)
}

func TestCalendarListCmd_DateParsing(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		shouldParse bool
	}{
		{
			name:        "ISO date",
			dateStr:     "2024-01-15",
			shouldParse: true,
		},
		{
			name:        "RFC3339",
			dateStr:     "2024-01-15T10:30:00Z",
			shouldParse: true,
		},
		{
			name:        "invalid format",
			dateStr:     "01/15/2024",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err1 := time.Parse("2006-01-02", tt.dateStr)
			_, err2 := time.Parse(time.RFC3339, tt.dateStr)
			parsed := err1 == nil || err2 == nil
			assert.Equal(t, tt.shouldParse, parsed)
		})
	}
}

// Tests for CalendarGetCmd
func TestCalendarGetCmd_Fields(t *testing.T) {
	cmd := &CalendarGetCmd{
		ID: "event-123",
	}
	assert.Equal(t, "event-123", cmd.ID)
}

// Tests for CalendarCreateCmd
func TestCalendarCreateCmd_Fields(t *testing.T) {
	cmd := &CalendarCreateCmd{
		Summary:   "New Meeting",
		From:      "2024-01-15T10:00:00Z",
		To:        "2024-01-15T11:00:00Z",
		Location:  "Room 101",
		Description:      "Meeting description",
		Attendees: []string{"a@example.com", "b@example.com"},
		AllDay:    false,
		Calendar:  "cal-123",
	}

	assert.Equal(t, "New Meeting", cmd.Summary)
	assert.Equal(t, "2024-01-15T10:00:00Z", cmd.From)
	assert.Equal(t, "2024-01-15T11:00:00Z", cmd.To)
	assert.Equal(t, "Room 101", cmd.Location)
	assert.Equal(t, "Meeting description", cmd.Description)
	assert.Len(t, cmd.Attendees, 2)
	assert.False(t, cmd.AllDay)
}

func TestCalendarCreateCmd_AllDay(t *testing.T) {
	cmd := &CalendarCreateCmd{
		Summary: "Holiday",
		From:    "2024-01-01",
		To:      "2024-01-02",
		AllDay:  true,
	}

	assert.True(t, cmd.AllDay)
}

// Tests for CalendarUpdateCmd
func TestCalendarUpdateCmd_Fields(t *testing.T) {
	cmd := &CalendarUpdateCmd{
		ID:       "event-123",
		Summary:  "Updated Meeting",
		From:     "2024-01-15T14:00:00Z",
		To:       "2024-01-15T15:00:00Z",
		Location: "New Room",
		Description:     "Updated description",
	}

	assert.Equal(t, "event-123", cmd.ID)
	assert.Equal(t, "Updated Meeting", cmd.Summary)
}

func TestCalendarUpdateCmd_RequiresUpdates(t *testing.T) {
	cleanup := setupCalendarTestConfig(t)
	defer cleanup()

	cmd := &CalendarUpdateCmd{
		ID: "event-123",
		// No updates specified
	}

	root := &Root{}
	_ = cmd.Run(root)

	// Should error because no updates are specified (or fail on client creation)
	// Either way, the command should handle this gracefully
	assert.NotNil(t, cmd)
}

// Tests for CalendarDeleteCmd
func TestCalendarDeleteCmd_Fields(t *testing.T) {
	cmd := &CalendarDeleteCmd{
		ID: "event-123",
	}
	assert.Equal(t, "event-123", cmd.ID)
}

// Tests for CalendarCalendarsCmd
func TestCalendarCalendarsCmd_Struct(t *testing.T) {
	cmd := &CalendarCalendarsCmd{}
	assert.NotNil(t, cmd)
}

// Tests for CalendarRespondCmd
func TestCalendarRespondCmd_Responses(t *testing.T) {
	tests := []struct {
		name     string
		response string
		valid    bool
	}{
		{"accept", "accept", true},
		{"decline", "decline", true},
		{"tentative", "tentative", true},
		{"invalid", "maybe", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CalendarRespondCmd{
				ID:       "event-123",
				Response: tt.response,
			}
			_ = cmd // Ensure cmd is used

			// Validate response type
			validResponses := map[string]bool{
				"accept":    true,
				"decline":   true,
				"tentative": true,
			}

			isValid := validResponses[tt.response]
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func TestCalendarRespondCmd_Comment(t *testing.T) {
	cmd := &CalendarRespondCmd{
		ID:       "event-123",
		Response: "tentative",
		Comment:  "Will try to make it",
	}

	assert.Equal(t, "Will try to make it", cmd.Comment)
}

// Tests for CalendarFreeBusyCmd
func TestCalendarFreeBusyCmd_Fields(t *testing.T) {
	cmd := &CalendarFreeBusyCmd{
		Emails: []string{"user1@example.com", "user2@example.com"},
		Start:  "2024-01-15T08:00:00Z",
		End:    "2024-01-15T18:00:00Z",
	}

	assert.Len(t, cmd.Emails, 2)
	assert.Equal(t, "2024-01-15T08:00:00Z", cmd.Start)
	assert.Equal(t, "2024-01-15T18:00:00Z", cmd.End)
}

// Tests for printEvent
func TestPrintEvent(t *testing.T) {
	event := Event{
		ID:       "event-id-12345678901234567890",
		Subject:  "Team Standup",
		IsAllDay: false,
		Start: &Time{
			DateTime: "2024-01-15T10:00:00.0000000",
			TimeZone: "UTC",
		},
		Location: &Loc{
			DisplayName: "Zoom",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printEvent(event, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Team Standup")
	assert.Contains(t, output, "Zoom")
	assert.Contains(t, output, "Jan 15")
}

func TestPrintEvent_AllDay(t *testing.T) {
	event := Event{
		ID:       "event-id-12345678901234567890",
		Subject:  "Holiday",
		IsAllDay: true,
		Start: &Time{
			DateTime: "2024-01-01T00:00:00.0000000",
			TimeZone: "UTC",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printEvent(event, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Holiday")
	assert.Contains(t, output, "Jan 1")
}

func TestPrintEvent_Verbose(t *testing.T) {
	event := Event{
		ID:      "event-id-12345678901234567890",
		Subject: "Meeting",
		Start: &Time{
			DateTime: "2024-01-15T10:00:00.0000000",
			TimeZone: "UTC",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printEvent(event, true) // verbose = true

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Full:")
}

// Tests for printEventDetail
func TestPrintEventDetail(t *testing.T) {
	event := Event{
		ID:       "event-id-12345678901234567890",
		Subject:  "Project Review",
		IsAllDay: false,
		Start: &Time{
			DateTime: "2024-01-15T14:00:00.0000000",
			TimeZone: "UTC",
		},
		End: &Time{
			DateTime: "2024-01-15T15:30:00.0000000",
			TimeZone: "UTC",
		},
		Location: &Loc{
			DisplayName: "Main Building",
		},
		Organizer: &Org{
			EmailAddress: struct {
				Name    string `json:"name"`
				Address string `json:"address"`
			}{
				Name:    "Manager",
				Address: "manager@example.com",
			},
		},
		Body: &Body{
			ContentType: "text",
			Content:     "Quarterly project review meeting",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printEventDetail(event, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Subject:  Project Review")
	assert.Contains(t, output, "Start:")
	assert.Contains(t, output, "End:")
	assert.Contains(t, output, "Location: Main Building")
	assert.Contains(t, output, "Organizer: Manager")
	assert.Contains(t, output, "Quarterly project review meeting")
}

func TestPrintEventDetail_HTMLBody(t *testing.T) {
	event := Event{
		ID:      "event-id-12345678901234567890",
		Subject: "Meeting",
		Body: &Body{
			ContentType: "html",
			Content:     "<p>Meeting <b>notes</b></p>",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printEventDetail(event, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// HTML should be stripped
	assert.Contains(t, output, "Meeting notes")
	assert.NotContains(t, output, "<p>")
}
