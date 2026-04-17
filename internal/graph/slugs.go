// Package graph provides slug ID management.
package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/iskycc/mogcli/internal/config"
)

var (
	slugCache  *config.Slugs
	slugMu     sync.Mutex
	aliasCache *config.Aliases
	aliasMu    sync.Mutex
)

// FormatID converts a long Microsoft Graph ID to a short slug.
func FormatID(id string) string {
	if id == "" {
		return ""
	}

	slugMu.Lock()
	defer slugMu.Unlock()

	if slugCache == nil {
		var err error
		slugCache, err = config.LoadSlugs()
		if err != nil {
			slugCache = &config.Slugs{
				IDToSlug: make(map[string]string),
				SlugToID: make(map[string]string),
			}
		}
	}

	// Check if we already have a slug for this ID
	if slug, ok := slugCache.IDToSlug[id]; ok {
		return slug
	}

	// Generate a new slug
	hash := sha256.Sum256([]byte(id))
	slug := hex.EncodeToString(hash[:])[:8]

	// Handle collisions
	origSlug := slug
	counter := 0
	for {
		if existingID, ok := slugCache.SlugToID[slug]; !ok || existingID == id {
			break
		}
		counter++
		slug = origSlug[:6] + hex.EncodeToString([]byte{byte(counter)})[:2]
	}

	// Store the mapping
	slugCache.IDToSlug[id] = slug
	slugCache.SlugToID[slug] = id

	// Save to disk (ignore errors for performance)
	_ = config.SaveSlugs(slugCache)

	return slug
}

// ResolveID converts an alias, slug, or full ID back to a full ID.
// Resolution order: @alias → slug → full ID passthrough.
func ResolveID(input string) string {
	if input == "" {
		return ""
	}

	// Check for alias prefix
	if strings.HasPrefix(input, "@") {
		name := strings.TrimPrefix(input, "@")
		if target := resolveAlias(name); target != "" {
			// Target may be a slug — resolve it further
			return ResolveID(target)
		}
		// Unknown alias — return as-is so caller gets a clear error from the API
		return input
	}

	// If it looks like a full ID (long), return as-is
	if len(input) > 16 {
		return input
	}

	slugMu.Lock()
	defer slugMu.Unlock()

	if slugCache == nil {
		var err error
		slugCache, err = config.LoadSlugs()
		if err != nil {
			return input
		}
	}

	// Try to resolve as a slug
	if fullID, ok := slugCache.SlugToID[input]; ok {
		return fullID
	}

	// Return as-is (might be a short ID that we haven't seen)
	return input
}

// resolveAlias looks up an alias name in the cache.
func resolveAlias(name string) string {
	aliasMu.Lock()
	defer aliasMu.Unlock()

	if aliasCache == nil {
		var err error
		aliasCache, err = config.LoadAliases()
		if err != nil {
			return ""
		}
	}

	return aliasCache.NameToTarget[name]
}

// SetAlias creates or updates a named alias.
func SetAlias(name, target string) error {
	name = strings.TrimPrefix(name, "@")
	if name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}

	aliasMu.Lock()
	defer aliasMu.Unlock()

	if aliasCache == nil {
		var err error
		aliasCache, err = config.LoadAliases()
		if err != nil {
			aliasCache = &config.Aliases{NameToTarget: make(map[string]string)}
		}
	}

	aliasCache.NameToTarget[name] = target
	return config.SaveAliases(aliasCache)
}

// DeleteAlias removes a named alias.
func DeleteAlias(name string) error {
	name = strings.TrimPrefix(name, "@")

	aliasMu.Lock()
	defer aliasMu.Unlock()

	if aliasCache == nil {
		var err error
		aliasCache, err = config.LoadAliases()
		if err != nil {
			return err
		}
	}

	if _, ok := aliasCache.NameToTarget[name]; !ok {
		return fmt.Errorf("alias @%s not found", name)
	}

	delete(aliasCache.NameToTarget, name)
	return config.SaveAliases(aliasCache)
}

// ListAliases returns all aliases sorted by name.
func ListAliases() ([]AliasEntry, error) {
	aliasMu.Lock()
	defer aliasMu.Unlock()

	if aliasCache == nil {
		var err error
		aliasCache, err = config.LoadAliases()
		if err != nil {
			return nil, err
		}
	}

	entries := make([]AliasEntry, 0, len(aliasCache.NameToTarget))
	for name, target := range aliasCache.NameToTarget {
		entries = append(entries, AliasEntry{Name: name, Target: target})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

// GetAlias returns the target for a named alias.
func GetAlias(name string) (string, error) {
	name = strings.TrimPrefix(name, "@")

	aliasMu.Lock()
	defer aliasMu.Unlock()

	if aliasCache == nil {
		var err error
		aliasCache, err = config.LoadAliases()
		if err != nil {
			return "", err
		}
	}

	target, ok := aliasCache.NameToTarget[name]
	if !ok {
		return "", fmt.Errorf("alias @%s not found", name)
	}
	return target, nil
}

// ClearAliases clears the alias cache.
func ClearAliases() {
	aliasMu.Lock()
	defer aliasMu.Unlock()
	aliasCache = nil
}

// AliasEntry represents a single alias mapping.
type AliasEntry struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// ClearSlugs clears the slug cache.
func ClearSlugs() error {
	slugMu.Lock()
	defer slugMu.Unlock()

	slugCache = &config.Slugs{
		IDToSlug: make(map[string]string),
		SlugToID: make(map[string]string),
	}

	return config.SaveSlugs(slugCache)
}
