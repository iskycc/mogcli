package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/iskycc/mogcli/internal/config"
	"github.com/iskycc/mogcli/internal/graph"
)

// AuthCmd handles authentication commands.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Login to Microsoft 365"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
	Logout AuthLogoutCmd `cmd:"" help:"Logout and clear tokens"`
	List   AuthListCmd   `cmd:"" help:"List configured accounts"`
}

// AuthLoginCmd logs in to Microsoft 365.
type AuthLoginCmd struct {
	ClientID string `help:"Azure AD client ID" required:"" env:"MOG_CLIENT_ID" name:"client-id"`
	Storage  string `help:"Token storage: file or keychain" default:"file" enum:"file,keychain"`
	Region   string `help:"Azure AD region: global or china" default:"global" enum:"global,china" env:"MOG_REGION"`
}

// Run executes the auth login command.
func (c *AuthLoginCmd) Run(root *Root) error {
	// Set storage type
	if c.Storage == "keychain" {
		config.SetStorage(config.StorageKeyring)
	} else {
		config.SetStorage(config.StorageFile)
	}

	// Save client ID, storage preference, and region
	cfg := &config.Config{ClientID: c.ClientID, Storage: c.Storage, Region: c.Region}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Request device code
	fmt.Printf("Requesting device code for account '%s'...\n", config.GetAccount())
	dcResp, err := graph.RequestDeviceCode(c.ClientID)
	if err != nil {
		return fmt.Errorf("failed to request device code: %w", err)
	}

	fmt.Println()
	fmt.Println(dcResp.Message)
	fmt.Println()

	// Try to open browser
	openBrowser(dcResp.VerificationURI)

	// Poll for token
	fmt.Println("Waiting for authorization...")
	tokens, err := graph.PollForToken(c.ClientID, dcResp.DeviceCode, dcResp.Interval)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	// Save tokens using configured storage
	if err := config.SaveTokensAuto(tokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Successfully logged in as account '%s' (storage: %s)\n", config.GetAccount(), c.Storage)
	return nil
}

// AuthStatusCmd shows authentication status.
type AuthStatusCmd struct{}

// Run executes the auth status command.
func (c *AuthStatusCmd) Run(root *Root) error {
	// Load config to get storage preference
	cfg, _ := config.Load()
	if cfg != nil && cfg.Storage == "keychain" {
		config.SetStorage(config.StorageKeyring)
	}

	fmt.Printf("Account: %s\n", config.GetAccount())

	if cfg != nil {
		fmt.Printf("Region: %s\n", cfg.GetRegion())
	}

	tokens, err := config.LoadTokensAuto()
	if err != nil {
		fmt.Println("Status: Not logged in")
		return nil
	}

	fmt.Println("Status: Logged in")
	if cfg != nil && cfg.Storage != "" {
		fmt.Printf("Storage: %s\n", cfg.Storage)
	}

	if tokens.ExpiresAt > 0 {
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		remaining := time.Until(expiresAt)
		if remaining > 0 {
			fmt.Printf("Token expires: %s (in %v)\n", expiresAt.Format(time.RFC3339), remaining.Round(time.Minute))
		} else {
			fmt.Println("Token: Expired (will refresh on next request)")
		}
	}

	if cfg != nil && cfg.ClientID != "" {
		fmt.Printf("Client ID: %s...%s\n", cfg.ClientID[:8], cfg.ClientID[len(cfg.ClientID)-4:])
	}

	return nil
}

// AuthLogoutCmd logs out and clears tokens.
type AuthLogoutCmd struct{}

// Run executes the auth logout command.
func (c *AuthLogoutCmd) Run(root *Root) error {
	// Load config to get storage preference
	cfg, _ := config.Load()
	if cfg != nil && cfg.Storage == "keychain" {
		config.SetStorage(config.StorageKeyring)
	}

	if err := config.DeleteTokensAuto(); err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	if err := graph.ClearSlugs(); err != nil {
		return fmt.Errorf("failed to clear slugs: %w", err)
	}

	fmt.Printf("✓ Logged out from account '%s' successfully\n", config.GetAccount())
	return nil
}

// AuthListCmd lists configured accounts.
type AuthListCmd struct{}

// Run executes the auth list command.
func (c *AuthListCmd) Run(root *Root) error {
	accounts, err := config.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts configured.")
		fmt.Println("Run: mog auth login --client-id YOUR_CLIENT_ID")
		return nil
	}

	currentAccount := config.GetAccount()
	fmt.Println("Configured accounts:")
	for _, account := range accounts {
		if account == currentAccount {
			fmt.Printf("  * %s (current)\n", account)
		} else {
			fmt.Printf("    %s\n", account)
		}
	}

	fmt.Println()
	fmt.Println("Use --account NAME or MOG_ACCOUNT=NAME to switch accounts.")
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
