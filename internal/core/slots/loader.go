// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
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
		return nil, fmt.Errorf("errors.slots.load_failed: %s: %w", tomlDir, err)
	}

	var items []SlotItem
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		filePath := filepath.Join(tomlDir, entry.Name())
		content, err := embeddedTOMLs.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("errors.slots.load_failed: %s: %w", filePath, err)
		}

		var slot tomlSlot
		if err := toml.Unmarshal(content, &slot); err != nil {
			return nil, fmt.Errorf("errors.slots.load_failed: %s: %w", filePath, err)
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
// NOTE: Esta función usa el filesystem adapter para seguir las Golden Rules.
func LoadSlotsFromTOML(fs ports.IFileSystem, basePath string) ([]SlotItem, error) {
	var items []SlotItem

	// Walk the directory tree usando el puerto IFileSystem
	err := fs.WalkDir(basePath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // Skip errors and continue
		}

		// Skip directories and non-.toml files
		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		// Parse the TOML file
		slot, err := parseTOML(path)
		if err != nil {
			return fmt.Errorf("errors.slots.load_failed: %s: %w", path, err)
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
// It always scans the filesystem to support dynamic category discovery.
// The axiomPath parameter is the root directory of the AXIOM project.
func LoadAndRegisterSlots(axiomPath string) error {
	fs := filesystem.NewFSAdapter()
	return loadFromFilesystem(fs, axiomPath)
}

// loadFromFilesystem loads slot items from TOML files on the filesystem.
// This is used as a fallback when embedded files are not available.
// It dynamically discovers all categories by scanning subdirectories.
// The axiomPath parameter is the root directory of the AXIOM project.
func loadFromFilesystem(fs ports.IFileSystem, axiomPath string) error {
	basePath := filepath.Join(axiomPath, "internal", "core", "slots")

	// Verify the base path exists
	if _, err := fs.Stat(basePath); err != nil {
		return fmt.Errorf("errors.slots.slots_not_found: %s", basePath)
	}

	// Walk through all subdirectories looking for "tomls" folders
	err := fs.WalkDir(basePath, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Look for directories named "tomls"
		if info.IsDir() && strings.HasSuffix(path, "tomls") {
			items, err := LoadSlotsFromTOML(fs, path)
			if err != nil {
				// Return error instead of printing to stderr (Regla 2 — el core es mudo)
				return fmt.Errorf("errors.slots.load_failed: %s: %w", path, err)
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
