// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// tomlSlot represents the structure of a slot TOML file.
type tomlSlot struct {
	ID           string   `toml:"id"`
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	Category     string   `toml:"category"`
	SubCategory  string   `toml:"subcategory"`
	Dependencies []string `toml:"dependencies"`
	Install      struct {
		Cmd   string   `toml:"cmd"`
		Steps []string `toml:"steps"`
	} `toml:"install"`
}

// LoadSlotsFromTOML loads all slot items from TOML files in the given directory.
// It recursively searches for .toml files and creates SlotItem instances.
func LoadSlotsFromTOML(basePath string) ([]SlotItem, error) {
	var items []SlotItem

	// Walk the directory tree
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.toml files
		if info.IsDir() || filepath.Ext(path) != ".toml" {
			return nil
		}

		// Parse the TOML file
		slot, err := parseTOML(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Convert to SlotItem
		item := SlotItem{
			ID:           slot.ID,
			Name:         slot.Name,
			Description:  slot.Description,
			Category:     SlotCategory(slot.Category),
			SubCategory:  slot.SubCategory,
			Deps:         slot.Dependencies,
			InstallCmd:   slot.Install.Cmd,
			InstallSteps: slot.Install.Steps,
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

// parseTOML parses a single TOML file into a tomlSlot struct.
func parseTOML(path string) (*tomlSlot, error) {
	var slot tomlSlot
	if _, err := toml.DecodeFile(path, &slot); err != nil {
		return nil, err
	}
	return &slot, nil
}

// LoadAndRegisterSlots loads slots from TOML files and registers them in the global registry.
// It looks for TOML files in internal/slots/**/toml/*.toml patterns.
func LoadAndRegisterSlots() error {
	// Get the project root (assume we're in internal/slots)
	basePath := "."

	// Try to find TOML directories
	tomlPaths := []string{
		"internal/slots/dev/ia/toml",
		"internal/slots/dev/languages/toml",
		"internal/slots/dev/tools/toml",
		"internal/slots/data/toml",
		"internal/slots/sandbox/toml",
	}

	for _, tomlPath := range tomlPaths {
		fullPath := filepath.Join(basePath, tomlPath)
		items, err := LoadSlotsFromTOML(fullPath)
		if err != nil {
			// Skip if directory doesn't exist
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to load slots from %s: %w", tomlPath, err)
		}

		// Register each item
		for i := range items {
			RegisterItem(&items[i])
		}
	}

	return nil
}
