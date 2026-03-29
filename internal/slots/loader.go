// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed dev/ia/tomls/*.toml
//go:embed dev/languages/tomls/*.toml
//go:embed dev/tools/tomls/*.toml
//go:embed data/tomls/*.toml
//go:embed sandbox/tomls/*.toml
var embeddedTOMLs embed.FS

// tomlSlot represents the structure of a slot TOML file.
type tomlSlot struct {
	ID           string   `toml:"id"`
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	Category     string   `toml:"category"`
	SubCategory  string   `toml:"subcategory"`
	Dependencies []string `toml:"dependencies"`
	IsBaseTool   bool     `toml:"is_base_tool"`
	Install      struct {
		Cmd   string   `toml:"cmd"`
		Steps []string `toml:"steps"`
	} `toml:"install"`
}

// LoadSlotsFromEmbeddedTOML loads all slot items from embedded TOML files in the given directory path.
// The path should be relative to the internal/slots directory (e.g., "dev/ia/tomls").
func LoadSlotsFromEmbeddedTOML(tomlDir string) ([]SlotItem, error) {
	entries, err := embeddedTOMLs.ReadDir(tomlDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded directory %s: %w", tomlDir, err)
	}

	var items []SlotItem
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		filePath := filepath.Join(tomlDir, entry.Name())
		content, err := embeddedTOMLs.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded file %s: %w", filePath, err)
		}

		var slot tomlSlot
		if err := toml.Unmarshal(content, &slot); err != nil {
			return nil, fmt.Errorf("failed to parse embedded TOML %s: %w", filePath, err)
		}

		item := SlotItem{
			ID:           slot.ID,
			Name:         slot.Name,
			Description:  slot.Description,
			Category:     SlotCategory(slot.Category),
			SubCategory:  slot.SubCategory,
			Deps:         slot.Dependencies,
			IsBaseTool:   slot.IsBaseTool,
			InstallCmd:   slot.Install.Cmd,
			InstallSteps: slot.Install.Steps,
		}

		items = append(items, item)
	}

	return items, nil
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
			IsBaseTool:   slot.IsBaseTool,
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
// If items are already registered (via init() functions), it skips loading.
// Otherwise, it first tries to load from embedded files (for production builds),
// then falls back to filesystem loading (for development).
func LoadAndRegisterSlots() error {
	// Check if items are already registered (from init() functions)
	if ItemCount() > 0 {
		return nil
	}

	// Track if we loaded anything
	loadedCount := 0

	// List of embedded TOML directories to try
	tomlDirs := []string{
		"dev/ia/tomls",
		"dev/languages/tomls",
		"dev/tools/tomls",
		"data/tomls",
		"sandbox/tomls",
	}

	// Try to load from embedded files first
	embedWorks := true
	for _, tomlDir := range tomlDirs {
		items, err := LoadSlotsFromEmbeddedTOML(tomlDir)
		if err != nil {
			embedWorks = false
			break
		}
		for i := range items {
			RegisterItem(&items[i])
			loadedCount++
		}
	}

	// If embedded loading worked and we got items, we're done
	if embedWorks && loadedCount > 0 {
		return nil
	}

	// Fallback to filesystem loading for development
	return loadFromFilesystem()
}

// loadFromFilesystem loads slot items from TOML files on the filesystem.
// This is used as a fallback when embedded files are not available.
func loadFromFilesystem() error {
	// Get AXIOM root from environment or use relative path
	rootDir := os.Getenv("AXIOM_PATH")
	if rootDir == "" {
		// Try to find project root from current directory
		rootDir = findProjectRoot()
	}

	basePath := filepath.Join(rootDir, "internal", "slots")

	// Verify the base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("slots directory not found: %s", basePath)
	}

	// Walk through all subdirectories looking for "tomls" folders
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Look for directories named "tomls"
		if info.IsDir() && strings.HasSuffix(path, "tomls") {
			items, err := LoadSlotsFromTOML(path)
			if err != nil {
				// Log warning but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to load slots from %s: %v\n", path, err)
				return nil
			}

			// Register each item
			for i := range items {
				RegisterItem(&items[i])
			}
		}

		return nil
	})

	return err
}

// findProjectRoot attempts to find the project root directory
// by looking for go.mod or internal/slots directory
func findProjectRoot() string {
	// Try current directory
	if _, err := os.Stat("go.mod"); err == nil {
		return "."
	}

	// Try parent directory
	if _, err := os.Stat("../go.mod"); err == nil {
		return ".."
	}

	// Try internal/slots relative path
	if _, err := os.Stat("internal/slots"); err == nil {
		return "."
	}

	// Default to current directory
	return "."
}
