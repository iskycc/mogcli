package cli

import (
	"encoding/json"
	"fmt"

	"github.com/visionik/mogcli/internal/graph"
)

// AliasCmd handles alias operations.
type AliasCmd struct {
	Set  AliasSetCmd  `cmd:"" help:"Create or update an alias"`
	Rm   AliasRmCmd   `cmd:"" help:"Remove an alias"`
	List AliasListCmd `cmd:"" help:"List all aliases"`
	Get  AliasGetCmd  `cmd:"" help:"Show what an alias resolves to"`
}

// AliasSetCmd creates or updates an alias.
type AliasSetCmd struct {
	Name   string `arg:"" help:"Alias name (with or without @ prefix)"`
	Target string `arg:"" help:"Target slug or full ID"`
}

// Run executes alias set.
func (c *AliasSetCmd) Run(root *Root) error {
	if err := graph.SetAlias(c.Name, c.Target); err != nil {
		return err
	}

	name := c.Name
	if len(name) > 0 && name[0] != '@' {
		name = "@" + name
	}
	fmt.Printf("✓ Alias %s → %s\n", name, c.Target)
	return nil
}

// AliasRmCmd removes an alias.
type AliasRmCmd struct {
	Name string `arg:"" help:"Alias name (with or without @ prefix)"`
}

// Run executes alias rm.
func (c *AliasRmCmd) Run(root *Root) error {
	if err := graph.DeleteAlias(c.Name); err != nil {
		return err
	}

	name := c.Name
	if len(name) > 0 && name[0] != '@' {
		name = "@" + name
	}
	fmt.Printf("✓ Alias %s removed\n", name)
	return nil
}

// AliasListCmd lists all aliases.
type AliasListCmd struct{}

// Run executes alias list.
func (c *AliasListCmd) Run(root *Root) error {
	entries, err := graph.ListAliases()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No aliases configured.")
		fmt.Println("Create one: mog alias set @name <slug-or-id>")
		return nil
	}

	if root.JSON {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	for _, e := range entries {
		fmt.Printf("@%-20s → %s\n", e.Name, e.Target)
	}
	return nil
}

// AliasGetCmd shows what an alias resolves to.
type AliasGetCmd struct {
	Name string `arg:"" help:"Alias name (with or without @ prefix)"`
}

// Run executes alias get.
func (c *AliasGetCmd) Run(root *Root) error {
	target, err := graph.GetAlias(c.Name)
	if err != nil {
		return err
	}

	name := c.Name
	if len(name) > 0 && name[0] != '@' {
		name = "@" + name
	}

	if root.JSON {
		data, err := json.MarshalIndent(map[string]string{
			"name":   name,
			"target": target,
		}, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("@%s → %s\n", c.Name, target)

	// Also show what it ultimately resolves to if target is a slug
	resolved := graph.ResolveID(target)
	if resolved != target {
		fmt.Printf("  (resolves to: %s)\n", resolved)
	}

	return nil
}
