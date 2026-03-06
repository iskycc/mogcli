// Package cli defines the command-line interface for mog.
package cli

import (
	"fmt"
	"os"

	"github.com/visionik/mogcli/internal/config"
	"github.com/visionik/mogcli/internal/graph"
)

// ClientFactory is a function that creates a Graph client.
// This allows dependency injection for testing.
type ClientFactory func() (graph.Client, error)

// Root is the top-level CLI structure.
type Root struct {
	// Global flags
	Account string      `help:"Account name for multi-account support" env:"MOG_ACCOUNT" default:"default" short:"a"`
	AIHelp  bool        `name:"ai-help" help:"Show detailed help for AI/LLM agents"`
	JSON    bool        `help:"Output JSON to stdout (best for scripting)" xor:"format"`
	Plain   bool        `help:"Output stable, parseable text to stdout (TSV; no colors)" xor:"format"`
	Verbose bool        `help:"Show full IDs and extra details" short:"v"`
	Force   bool        `help:"Skip confirmations for destructive commands"`
	NoInput bool        `help:"Never prompt; fail instead (useful for CI)" name:"no-input"`
	Version VersionFlag `name:"version" help:"Print version and exit"`

	// Subcommands
	Auth     AuthCmd     `cmd:"" help:"Authentication"`
	Mail     MailCmd     `cmd:"" aliases:"email" help:"Mail operations"`
	Calendar CalendarCmd `cmd:"" aliases:"cal" help:"Calendar operations"`
	Drive    DriveCmd    `cmd:"" help:"OneDrive file operations"`
	Contacts ContactsCmd `cmd:"" help:"Contact operations"`
	Tasks    TasksCmd    `cmd:"" aliases:"todo" help:"Microsoft To-Do tasks"`
	Excel    ExcelCmd    `cmd:"" help:"Excel spreadsheet operations"`
	OneNote  OneNoteCmd  `cmd:"" aliases:"onenote" help:"OneNote operations"`
	Word     WordCmd     `cmd:"" help:"Word document operations"`
	PPT      PPTCmd      `cmd:"" aliases:"ppt,powerpoint" help:"PowerPoint operations"`
	Alias    AliasCmd    `cmd:"" help:"Manage named aliases for IDs and slugs"`

	// ClientFactory allows injecting a custom client factory for testing.
	// If nil, graph.NewClient is used.
	ClientFactory ClientFactory `kong:"-"`
}

// AfterApply runs after flags are parsed but before the command executes.
// This sets up the account context for all commands.
func (r *Root) AfterApply() error {
	// Set the active account
	config.SetAccount(r.Account)

	// Migrate legacy config if needed (first run after upgrade)
	if err := config.MigrateIfNeeded(); err != nil {
		// Non-fatal, just log if verbose
		if r.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: migration check failed: %v\n", err)
		}
	}

	return nil
}

// GetClient returns a Graph client using the configured factory or default.
func (r *Root) GetClient() (graph.Client, error) {
	if r.ClientFactory != nil {
		return r.ClientFactory()
	}
	return graph.NewClient()
}

// VersionFlag handles --version.
type VersionFlag string

// BeforeApply prints version and exits.
func (v VersionFlag) BeforeApply() error {
	fmt.Println(v)
	os.Exit(0)
	return nil
}
