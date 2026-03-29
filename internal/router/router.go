// Package router handles command routing from CLI to domain handlers.
package router

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"axiom/internal/domain"
	"axiom/internal/ports"
	"axiom/internal/slots"
)

// Command constants.
const (
	CmdCreate      = "create"
	CmdDelete      = "delete"
	CmdList        = "list"
	CmdStop        = "stop"
	CmdPrune       = "prune"
	CmdBuild       = "build"
	CmdRebuild     = "rebuild"
	CmdHelp        = "help"
	CmdInfo        = "info"
	CmdReset       = "reset"
	CmdEnter       = "enter"
	CmdInit        = "init"
	CmdDeleteImage = "delete-image"
	CmdSlots       = "slots"
)

// BunkerManagerInterface defines the contract for bunker operations.
type BunkerManagerInterface interface {
	Create(name string) error
	CreateWithImage(name, image string) error
	Delete(name string) error
	List() error
	Stop() error
	Prune() error
	Info(name string) error
	DeleteImage() error
	Help() error
	GetUI() ports.IPresenter
	LoadConfig() (domain.EnvConfig, error)
}

// BuildManagerInterface defines the contract for build operations.
type BuildManagerInterface interface {
	Build(ctx context.Context, cfg domain.EnvConfig) error
	Rebuild(ctx context.Context, cfg domain.EnvConfig) error
}

// SlotManagerInterface defines the contract for slot operations.
type SlotManagerInterface interface {
	DiscoverSlots() []any
	GetAvailableItems(category string) ([]slots.SlotItem, error)
	ExecuteSlots(selected []any) error
	HasSelection() bool
	GetSelectedItems(category string) ([]slots.SlotItem, error)
	RunSlotSelector(category string, items []slots.SlotItem, preselected []string) ([]string, bool, error)
	SaveSelection(selections []slots.SlotSelection) error
	LoadSelection() ([]slots.SlotSelection, error)
}

// Router dispatches CLI commands to appropriate managers.
type Router struct {
	bm  BunkerManagerInterface
	bld BuildManagerInterface
	slm SlotManagerInterface
}

// NewRouter creates a new Router with all managers.
func NewRouter(bm BunkerManagerInterface, bld BuildManagerInterface, slm SlotManagerInterface) *Router {
	return &Router{
		bm:  bm,
		bld: bld,
		slm: slm,
	}
}

var knownCommands = map[string]struct{}{
	CmdCreate:      {},
	CmdDelete:      {},
	CmdList:        {},
	CmdStop:        {},
	CmdPrune:       {},
	CmdBuild:       {},
	CmdRebuild:     {},
	CmdHelp:        {},
	CmdInfo:        {},
	CmdReset:       {},
	CmdEnter:       {},
	CmdInit:        {},
	CmdDeleteImage: {},
	CmdSlots:       {},
	"rm":           {},
	"ls":           {},
	"-h":           {},
	"--help":       {},
}

// KnownCommand returns true if the given command is recognized.
func KnownCommand(cmd string) bool {
	_, ok := knownCommands[strings.ToLower(cmd)]
	return ok
}

// Handle routes commands to appropriate managers.
func (r *Router) Handle(args []string) error {
	if len(args) == 0 {
		return r.bm.Help()
	}

	cmd := normalizeCommand(args[0])
	if cmd == "" {
		return r.bm.Help()
	}

	firstArg := firstArg(args)

	switch cmd {
	case CmdCreate:
		return r.handleCreate()
	case CmdDelete, "rm":
		return r.bm.Delete(firstArg)
	case CmdList, "ls":
		return r.bm.List()
	case CmdStop:
		return r.bm.Stop()
	case CmdPrune:
		return r.bm.Prune()
	case CmdInfo:
		return r.bm.Info(firstArg)
	case CmdDeleteImage:
		return r.bm.DeleteImage()
	case CmdHelp, "-h", "--help":
		return r.bm.Help()
	case CmdBuild:
		// Load configuration
		cfg, err := r.bm.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Check if slot selection exists
		if !r.slm.HasSelection() {
			r.bm.GetUI().ShowLog("info", "No slot selection found. Running slot selector...")

			// Get ALL available items for DEV slot (not selected ones!)
			items, err := r.slm.GetAvailableItems("dev")
			if err != nil {
				return fmt.Errorf("failed to get slot items: %w", err)
			}

			// Run slot selector with available items
			selectedIDs, confirmed, err := r.slm.RunSlotSelector("dev", items, nil)
			if err != nil {
				return fmt.Errorf("slot selector failed: %w", err)
			}
			if !confirmed {
				return nil // User cancelled
			}

			// Save selection
			selections := []slots.SlotSelection{{Slot: slots.SlotDEV, Selected: selectedIDs}}
			if err := r.slm.SaveSelection(selections); err != nil {
				return fmt.Errorf("failed to save slot selection: %w", err)
			}
		}

		// Execute build with slot integration
		return r.bld.Build(context.Background(), cfg)
	case CmdRebuild:
		r.bm.GetUI().ShowLog("build.not_implemented")
		return nil
	case CmdReset:
		r.bm.GetUI().ShowLog("reset.not_implemented")
		return nil
	case CmdInit:
		r.bm.GetUI().ShowLog("system.init_not_implemented")
		return nil
	case CmdSlots:
		items := r.slm.DiscoverSlots()
		r.bm.GetUI().ShowLog("slots.discovered", len(items))
		return nil
	case CmdEnter:
		if firstArg == "" {
			return errors.New("usage: axiom enter <bunker-name>")
		}
		return errors.New("enter not yet implemented")
	default:
		return &unknownCommandError{cmd: cmd}
	}
}

func normalizeCommand(cmd string) string {
	return strings.ToLower(strings.TrimSpace(cmd))
}

func firstArg(args []string) string {
	if len(args) < 2 {
		return ""
	}
	return strings.TrimSpace(args[1])
}

type unknownCommandError struct {
	cmd string
}

func (e *unknownCommandError) Error() string {
	return "unknown command: " + e.cmd
}

// handleCreate implements the create command flow with image/slot selection.
func (r *Router) handleCreate() error {
	ui := r.bm.GetUI()

	// Image/slot options
	images := []string{
		"axiom-dev",
		"axiom-data",
		"axiom-sandbox",
	}

	// Show selection menu
	ui.ShowLogo()
	selected, err := ui.AskString("create.select_image")
	if err != nil {
		return fmt.Errorf("create.image_selection: %w", err)
	}

	selected = strings.TrimSpace(strings.ToLower(selected))

	// Resolve slot to image mapping or use exact name if provided
	imageName := resolveImageName(selected, images)
	if imageName == "" {
		return fmt.Errorf("create.invalid_image")
	}

	// Ask for bunker name
	bunkerName, err := ui.AskString("create.name_prompt")
	if err != nil {
		return fmt.Errorf("create.name_input: %w", err)
	}

	bunkerName = strings.TrimSpace(bunkerName)
	if bunkerName == "" {
		return fmt.Errorf("create.missing_name")
	}

	return r.bm.CreateWithImage(bunkerName, imageName)
}

// resolveImageName maps slot name to image or validates exact image name.
// Returns empty string if invalid.
func resolveImageName(input string, available []string) string {
	// Check for exact match in available images
	for _, img := range available {
		if strings.EqualFold(input, img) {
			return img
		}
	}

	// Map slot names to images
	slotMapping := map[string]string{
		"dev":     "axiom-dev",
		"data":    "axiom-data",
		"sandbox": "axiom-sandbox",
	}

	if mapped, ok := slotMapping[input]; ok {
		return mapped
	}

	// Check if input matches any available image by base name
	for _, img := range available {
		base := strings.TrimPrefix(img, "axiom-")
		if strings.EqualFold(input, base) {
			return img
		}
	}

	return ""
}
