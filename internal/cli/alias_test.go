package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/iskycc/mogcli/internal/graph"
)

// resetAliasCache clears the alias cache for test isolation.
func resetAliasCache(t *testing.T) {
	t.Helper()
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	// Clear in-memory cache so it reloads from the new HOME
	graph.ClearAliases()
}

func TestAliasSetCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *AliasSetCmd
		wantErr   bool
		wantInOut string
	}{
		{
			name:      "set alias without prefix",
			cmd:       &AliasSetCmd{Name: "standup", Target: "a3f2c891"},
			wantInOut: "✓ Alias @standup → a3f2c891",
		},
		{
			name:      "set alias with prefix",
			cmd:       &AliasSetCmd{Name: "@budget", Target: "b4c7d2e1"},
			wantInOut: "✓ Alias @budget → b4c7d2e1",
		},
		{
			name:    "set alias with empty name",
			cmd:     &AliasSetCmd{Name: "@", Target: "target"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasCache(t)
			root := &Root{}

			var output string
			var err error
			output = captureOutput(func() {
				err = tt.cmd.Run(root)
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, tt.wantInOut)
			}
		})
	}
}

func TestAliasRmCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		setup     string // alias name to create before test
		cmd       *AliasRmCmd
		wantErr   bool
		wantInOut string
	}{
		{
			name:      "remove existing alias",
			setup:     "todelete",
			cmd:       &AliasRmCmd{Name: "todelete"},
			wantInOut: "✓ Alias @todelete removed",
		},
		{
			name:      "remove with @ prefix",
			setup:     "prefixed",
			cmd:       &AliasRmCmd{Name: "@prefixed"},
			wantInOut: "✓ Alias @prefixed removed",
		},
		{
			name:    "remove nonexistent alias",
			cmd:     &AliasRmCmd{Name: "nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasCache(t)
			if tt.setup != "" {
				require.NoError(t, graph.SetAlias(tt.setup, "target"))
			}
			root := &Root{}

			var output string
			var err error
			output = captureOutput(func() {
				err = tt.cmd.Run(root)
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, tt.wantInOut)
			}
		})
	}
}

func TestAliasListCmd_Run(t *testing.T) {
	t.Run("list with aliases", func(t *testing.T) {
		resetAliasCache(t)
		require.NoError(t, graph.SetAlias("alpha", "target-a"))
		require.NoError(t, graph.SetAlias("beta", "target-b"))

		root := &Root{}
		output := captureOutput(func() {
			err := (&AliasListCmd{}).Run(root)
			require.NoError(t, err)
		})

		assert.Contains(t, output, "@alpha")
		assert.Contains(t, output, "@beta")
		assert.Contains(t, output, "target-a")
		assert.Contains(t, output, "target-b")
	})

	t.Run("list empty", func(t *testing.T) {
		resetAliasCache(t)

		root := &Root{}
		output := captureOutput(func() {
			err := (&AliasListCmd{}).Run(root)
			require.NoError(t, err)
		})

		assert.Contains(t, output, "No aliases configured")
	})

	t.Run("list with JSON output", func(t *testing.T) {
		resetAliasCache(t)
		require.NoError(t, graph.SetAlias("test", "target-t"))

		root := &Root{JSON: true}
		output := captureOutput(func() {
			err := (&AliasListCmd{}).Run(root)
			require.NoError(t, err)
		})

		assert.Contains(t, output, `"name"`)
		assert.Contains(t, output, `"target"`)
	})
}

func TestAliasGetCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		setup     string // alias name to create
		cmd       *AliasGetCmd
		root      *Root
		wantErr   bool
		wantInOut string
	}{
		{
			name:      "get existing alias",
			setup:     "myalias",
			cmd:       &AliasGetCmd{Name: "myalias"},
			root:      &Root{},
			wantInOut: "@myalias",
		},
		{
			name:      "get with @ prefix",
			setup:     "prefixed",
			cmd:       &AliasGetCmd{Name: "@prefixed"},
			root:      &Root{},
			wantInOut: "@prefixed",
		},
		{
			name:    "get nonexistent alias",
			cmd:     &AliasGetCmd{Name: "missing"},
			root:    &Root{},
			wantErr: true,
		},
		{
			name:      "get with JSON output",
			setup:     "jsontest",
			cmd:       &AliasGetCmd{Name: "jsontest"},
			root:      &Root{JSON: true},
			wantInOut: `"target"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasCache(t)
			if tt.setup != "" {
				require.NoError(t, graph.SetAlias(tt.setup, "target-value"))
			}

			var output string
			var err error
			output = captureOutput(func() {
				err = tt.cmd.Run(tt.root)
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, tt.wantInOut)
			}
		})
	}
}
