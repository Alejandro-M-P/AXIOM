package ports

// ISlotUI defines the interface for slot selection UI operations.
// Implementations can be TUI, CLI, or API.
type ISlotUI interface {
	// RunWizardWithSlot presents the wizard-style slot selector and returns both
	// selected item IDs and the selected slot (e.g., "dev", "data", "sandbox").
	// Returns (selectedIDs, selectedSlot, confirmed, error).
	RunWizardWithSlot(items any) ([]string, string, bool, error)
}
