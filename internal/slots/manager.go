// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"axiom/internal/ports"
)

// DefaultConfigPath is the default path for slot selection configuration.
const DefaultConfigPath = "configs/slots/selection.toml"

// SlotManager orchestrates slot discovery, selection, and execution.
// It uses dependency injection for its dependencies (registry, engine, presenter, filesystem).
type SlotManager struct {
	registry   ISlotRegistry
	engine     IInstallerEngine
	ui         ports.IPresenter
	fs         ports.IFileSystem
	configPath string
	uiRunner   UISelectorRunner
}

// UISelectorRunner defines the interface for running the slot selector UI.
// This allows the SlotManager to delegate UI operations while maintaining
// separation of concerns - the domain layer doesn't depend on UI directly.
type UISelectorRunner interface {
	RunSlotSelector(groups []ItemGroup) ([]string, error)
}

// ItemGroup represents a group of items for UI selection.
type ItemGroup struct {
	Title string
	Items []SlotItemDisplay
}

// NewItemGroup creates a new ItemGroup with the given title and items.
func NewItemGroup(title string, items []SlotItemDisplay) ItemGroup {
	return ItemGroup{Title: title, Items: items}
}

// SlotItemDisplay represents a slot item for UI display.
type SlotItemDisplay struct {
	ID          string
	Name        string
	Description string
}

// NewSlotItemDisplay creates a new SlotItemDisplay.
func NewSlotItemDisplay(id, name, description string) SlotItemDisplay {
	return SlotItemDisplay{ID: id, Name: name, Description: description}
}

// NewSlotManager creates a new SlotManager with the given dependencies.
func NewSlotManager(
	registry ISlotRegistry,
	engine IInstallerEngine,
	ui ports.IPresenter,
	fs ports.IFileSystem,
) *SlotManager {
	return &SlotManager{
		registry:   registry,
		engine:     engine,
		ui:         ui,
		fs:         fs,
		configPath: DefaultConfigPath,
	}
}

// NewSlotManagerWithConfig creates a new SlotManager with custom config path.
func NewSlotManagerWithConfig(
	registry ISlotRegistry,
	engine IInstallerEngine,
	ui ports.IPresenter,
	fs ports.IFileSystem,
	configPath string,
) *SlotManager {
	return &SlotManager{
		registry:   registry,
		engine:     engine,
		ui:         ui,
		fs:         fs,
		configPath: configPath,
	}
}

// Discover returns all available slot items from the registry.
func (m *SlotManager) Discover() []SlotItem {
	return m.registry.Discover()
}

// GetUI returns the presenter for output.
func (m *SlotManager) GetUI() ports.IPresenter {
	return m.ui
}

// DiscoverSlots returns all available slot items as generic slice.
func (m *SlotManager) DiscoverSlots() []any {
	items := m.registry.Discover()
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}

// GetByCategory returns all items for the specified slot category.
func (m *SlotManager) GetByCategory(category SlotCategory) []SlotItem {
	return m.registry.GetByCategory(category)
}

// GetAvailableItems returns all items for a category given as string.
// This implements the router.SlotManagerInterface.
func (m *SlotManager) GetAvailableItems(category string) ([]SlotItem, error) {
	return m.GetByCategory(SlotCategory(category)), nil
}

// GetAllAvailableItems returns ALL items from ALL categories.
// This is used by the build command to let users choose both slot and items.
func (m *SlotManager) GetAllAvailableItems() ([]SlotItem, error) {
	return m.registry.Discover(), nil
}

// Select returns slot items based on user configuration from the config file.
// If no config exists, it returns an empty selection.
func (m *SlotManager) Select(category SlotCategory) ([]SlotItem, error) {
	// Load saved selections
	selections, err := m.LoadSelection()
	if err != nil {
		// Log warning but continue with empty selection
		m.ui.ShowLog("warn", "Failed to load slot selection, using empty selection:", err.Error())
		selections = []SlotSelection{}
	}

	// Find selection for the requested category
	var selection SlotSelection
	found := false
	for _, sel := range selections {
		if sel.Slot == category {
			selection = sel
			found = true
			break
		}
	}

	if !found {
		return []SlotItem{}, nil
	}

	// Resolve selected item IDs to SlotItem objects
	var selectedItems []SlotItem
	for _, id := range selection.Selected {
		item, err := m.registry.GetByID(id)
		if err != nil {
			m.ui.ShowLog("warn", "Selected item not found:", id)
			continue
		}
		selectedItems = append(selectedItems, *item)
	}

	return selectedItems, nil
}

// HasSelection returns true if there are any saved slot selections.
func (m *SlotManager) HasSelection() bool {
	selections, err := m.LoadSelection()
	if err != nil {
		return false
	}
	for _, sel := range selections {
		if len(sel.Selected) > 0 {
			return true
		}
	}
	return false
}

// GetSelectedItems returns items for the specified category based on user selection.
// This is similar to Select() but uses string instead of SlotCategory for interface compatibility.
func (m *SlotManager) GetSelectedItems(category string) ([]SlotItem, error) {
	return m.Select(SlotCategory(category))
}

// Execute installs the given slot items using the installer engine.
// It resolves dependencies and installs items in the correct order.
func (m *SlotManager) Execute(selected []SlotItem) error {
	if len(selected) == 0 {
		return nil
	}

	// Use the engine to execute installation with dependency resolution
	return m.engine.Execute(selected, func(ctx context.Context) error {
		// Progress callback - could be used for logging
		return nil
	})
}

// ExecuteSlots installs the given slot items (generic version).
// This implements the router.SlotManagerInterface.
func (m *SlotManager) ExecuteSlots(selected []any) error {
	// Convert generic []any back to []SlotItem
	items := make([]SlotItem, 0, len(selected))
	for _, item := range selected {
		if si, ok := item.(SlotItem); ok {
			items = append(items, si)
		}
	}
	return m.Execute(items)
}

// ExecuteWithContext installs items with context support for cancellation.
func (m *SlotManager) ExecuteWithContext(ctx context.Context, selected []SlotItem) error {
	if len(selected) == 0 {
		return nil
	}

	return m.engine.ExecuteWithContext(ctx, selected, func(ctx context.Context, item *SlotItem) error {
		m.ui.ShowLog("info", "Installing:", item.Name)
		// Execute installation commands from TOML
		return m.executeInstall(ctx, item)
	})
}

// executeInstall runs the installation commands for a single item from TOML config.
func (m *SlotManager) executeInstall(ctx context.Context, item *SlotItem) error {
	// Skip if no install commands defined
	if item.InstallCmd == "" && len(item.InstallSteps) == 0 {
		return nil
	}

	// Single command mode
	if item.InstallCmd != "" {
		m.ui.ShowLog("info", fmt.Sprintf("Running: %s", item.InstallCmd))
		return m.runCommand(ctx, item.InstallCmd)
	}

	// Multi-step mode
	for i, step := range item.InstallSteps {
		m.ui.ShowLog("info", fmt.Sprintf("Step %d/%d: %s", i+1, len(item.InstallSteps), step))
		if err := m.runCommand(ctx, step); err != nil {
			return err
		}
	}

	return nil
}

// runCommand executes a shell command with context support.
func (m *SlotManager) runCommand(ctx context.Context, cmd string) error {
	command := exec.CommandContext(ctx, "sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// SaveSelection persists the user's slot selections to the config file.
func (m *SlotManager) SaveSelection(selections []SlotSelection) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := m.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := marshalTOML(selections)
	if err != nil {
		return fmt.Errorf("failed to marshal selection: %w", err)
	}

	// Write to file
	if err := m.fs.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	m.ui.ShowLog("info", "Selection saved to:", m.configPath)
	return nil
}

// LoadSelection reads the user's slot selections from the config file.
func (m *SlotManager) LoadSelection() ([]SlotSelection, error) {
	// Check if file exists
	if !m.fs.Exists(m.configPath) {
		return []SlotSelection{}, nil
	}

	// Read file
	data, err := m.fs.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal from TOML
	return unmarshalTOML(data)
}

// GetConfigPath returns the current configuration file path.
func (m *SlotManager) GetConfigPath() string {
	return m.configPath
}

// SetConfigPath sets a custom configuration file path.
func (m *SlotManager) SetConfigPath(path string) {
	m.configPath = path
}

// SetUISelectorRunner sets the UI runner for slot selection.
// This allows the SlotManager to delegate UI operations while maintaining
// separation of concerns.
func (m *SlotManager) SetUISelectorRunner(runner UISelectorRunner) {
	m.uiRunner = runner
}

// RunSlotSelector presents the interactive slot selector UI and returns selected item IDs.
// It uses the injected UISelectorRunner to display the UI.
// Returns ([]string, bool, error) where bool indicates if user confirmed selection.
func (m *SlotManager) RunSlotSelector(category string, items []SlotItem, preselected []string) ([]string, bool, error) {
	if m.uiRunner == nil {
		return nil, false, fmt.Errorf("UISelectorRunner not set: cannot run slot selector UI")
	}

	// Convert SlotItem to ItemGroup for UI
	groups := m.buildItemGroupsForUI(items)

	// Run the selector
	selectedIDs, err := m.uiRunner.RunSlotSelector(groups)
	if err != nil {
		return nil, false, fmt.Errorf("slot selector failed: %w", err)
	}

	if selectedIDs == nil {
		return nil, false, nil // User canceled
	}

	return selectedIDs, true, nil
}

// buildItemGroupsForUI converts SlotItems to ItemGroups for the UI selector.
func (m *SlotManager) buildItemGroupsForUI(items []SlotItem) []ItemGroup {
	// Group items by subcategory
	groupsMap := make(map[string][]SlotItemDisplay)
	order := []string{"ia", "languages", "tools", "data"}

	for _, item := range items {
		display := SlotItemDisplay{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
		groupsMap[item.SubCategory] = append(groupsMap[item.SubCategory], display)
	}

	// Build ordered result
	var result []ItemGroup
	for _, subcategory := range order {
		if items, ok := groupsMap[subcategory]; ok {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{Title: title, Items: items})
		}
	}

	// Add any remaining subcategories not in predefined order
	for subcategory, items := range groupsMap {
		found := false
		for _, o := range order {
			if o == subcategory {
				found = true
				break
			}
		}
		if !found {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{Title: title, Items: items})
		}
	}

	return result
}

// getSubcategoryTitle returns a human-readable title for a subcategory.
func getSubcategoryTitle(subcategory string) string {
	switch subcategory {
	case "ia":
		return "AI / LLM Models"
	case "languages":
		return "Programming Languages"
	case "tools":
		return "Developer Tools"
	case "data":
		return "Data Stores"
	default:
		return subcategory
	}
}

// marshalTOML marshals slot selections to TOML format.
func marshalTOML(selections []SlotSelection) ([]byte, error) {
	var result bytes.Buffer
	encoder := toml.NewEncoder(&result)

	// Create a serializable structure
	type selection struct {
		Slot     string   `toml:"slot"`
		Selected []string `toml:"selected"`
	}

	type selectionsWrapper struct {
		Selections []selection `toml:"selection"`
	}

	data := make([]selection, len(selections))
	for i, sel := range selections {
		data[i] = selection{
			Slot:     string(sel.Slot),
			Selected: sel.Selected,
		}
	}

	wrapper := selectionsWrapper{Selections: data}
	if err := encoder.Encode(wrapper); err != nil {
		return nil, fmt.Errorf("failed to encode TOML: %w", err)
	}

	return result.Bytes(), nil
}

// unmarshalTOML unmarshals slot selections from TOML format.
func unmarshalTOML(data []byte) ([]SlotSelection, error) {
	if len(data) == 0 {
		return []SlotSelection{}, nil
	}

	type selection struct {
		Slot     string   `toml:"slot"`
		Selected []string `toml:"selected"`
	}

	type selectionsWrapper struct {
		Selections []selection `toml:"selection"`
	}

	wrapper := selectionsWrapper{}
	if err := toml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	dataList := wrapper.Selections
	result := make([]SlotSelection, len(dataList))
	for i, sel := range dataList {
		result[i] = SlotSelection{
			Slot:     SlotCategory(sel.Slot),
			Selected: sel.Selected,
		}
	}

	return result, nil
}

// EnsureConfigDir ensures the configuration directory exists.
func EnsureConfigDir(fs ports.IFileSystem, baseDir string) error {
	configDir := filepath.Join(baseDir, "configs", "slots")
	info, err := fs.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fs.MkdirAll(configDir, 0755)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("config path exists but is not a directory: %s", configDir)
	}
	return nil
}
