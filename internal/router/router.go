// Package router handles command routing from CLI to domain handlers.
package router

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui"
	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/core/build"
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
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
	LoadConfig() (config.EnvConfig, error)
}

// BuildManagerInterface defines the contract for build operations.
type BuildManagerInterface interface {
	Build(ctx context.Context, cfg config.EnvConfig) (*build.BuildPlan, error)
	Rebuild(ctx context.Context, cfg config.EnvConfig) (*build.BuildPlan, error)
	SaveSlotSelection(selectedSlot string, selectedIDs []string) error
	GetUI() ports.IPresenter
}

// SlotManagerInterface defines the contract for slot operations.
type SlotManagerInterface interface {
	DiscoverSlots() []any
	GetAvailableItems(category string) ([]slots.SlotItem, error)
	GetAllAvailableItems() ([]slots.SlotItem, error)
	ExecuteSlots(selected []any) error
	HasSelection() bool
	GetSelectedItems(category string) ([]slots.SlotItem, error)
	RunSlotSelector(category string, items []slots.SlotItem, preselected []string) ([]string, bool, error)
	SaveSelection(selections []slots.SlotSelection) error
	LoadSelection() ([]slots.SlotSelection, error)
}

// Router dispatches CLI commands to appropriate managers.
type Router struct {
	bm        BunkerManagerInterface
	bld       BuildManagerInterface
	slm       SlotManagerInterface
	slotUI    ports.ISlotUI
	axiomPath string
	fs        ports.IFileSystem
}

// NewRouter creates a new Router with all managers.
func NewRouter(bm BunkerManagerInterface, bld BuildManagerInterface, slm SlotManagerInterface, slotUI ports.ISlotUI, axiomPath string, fs ports.IFileSystem) *Router {
	return &Router{
		bm:        bm,
		bld:       bld,
		slm:       slm,
		slotUI:    slotUI,
		axiomPath: axiomPath,
		fs:        fs,
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
		return r.bm.GetUI().RunHelpTUI()
	case CmdBuild:
		// Load configuration
		cfg, err := r.bm.LoadConfig()
		if err != nil {
			return fmt.Errorf("errors.router.failed_load_config: %w", err)
		}

		// Get ALL available items from ALL slots (we need all to let user choose slot first)
		allItems, err := r.slm.GetAllAvailableItems()
		if err != nil {
			return fmt.Errorf("errors.router.failed_get_slot_items: %w", err)
		}

		// Run wizard selector - this lets user choose slot (dev/data/sandbox) AND items
		// The wizard returns both the selected slot AND the selected item IDs
		selectedIDs, selectedSlot, confirmed, err := r.slotUI.RunWizardWithSlot(allItems)
		if err != nil {
			return fmt.Errorf("errors.router.slot_selector_failed: %w", err)
		}
		if !confirmed {
			return nil // User cancelled
		}

		// Save selection for the build process via BuildManager
		if err := r.bld.SaveSlotSelection(selectedSlot, selectedIDs); err != nil {
			return fmt.Errorf("errors.router.failed_save_slot_selection: %w", err)
		}

		// Generate the build plan from the core
		plan, err := r.bld.Build(context.Background(), cfg)
		if err != nil {
			return err
		}

		// Execute the plan with UI progress tracking (adapter layer)
		return r.executeBuildPlan(context.Background(), plan)
	case CmdRebuild:
		r.bm.GetUI().ShowLog("build.not_implemented")
		return nil
	case CmdReset:
		r.bm.GetUI().ShowLog("reset.not_implemented")
		return nil
	case CmdInit:
		return r.handleInit()
	case CmdSlots:
		items := r.slm.DiscoverSlots()
		r.bm.GetUI().ShowLog("slots.discovered", len(items))
		return nil
	case CmdEnter:
		if firstArg == "" {
			return errors.New("errors.router.usage_enter")
		}
		return fmt.Errorf("errors.router.enter_not_implemented")
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

	// Use TUI form for interactive creation
	bunkerName, imageName, confirmed, err := ui.AskCreateBunker(images)
	if err != nil {
		return fmt.Errorf("errors.router.create_form_error: %w", err)
	}
	if !confirmed {
		return nil // User cancelled
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

// handleInit ejecuta el wizard de inicialización TUI
func (r *Router) handleInit() error {
	// Verificar si existe archivo config.toml
	configPath := filepath.Join(r.axiomPath, "config.toml")
	configExists := r.fs.Exists(configPath)

	// Determinar idioma desde LANG o default "es"
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = "es"
	} else {
		// Extraer código de idioma base (ej: "en_US.UTF-8" -> "en")
		lang = strings.Split(lang, "_")[0]
		lang = strings.Split(lang, ".")[0]
		if lang != "en" && lang != "es" {
			lang = "es"
		}
	}

	// Obtener homeDir para el wizard
	homeDir, _ := r.fs.UserHomeDir()

	// Ejecutar el wizard TUI a través del presenter
	completed, err := r.bm.GetUI().RunInitWizardWithParams(context.Background(), r.axiomPath, configExists, lang, homeDir)
	if err != nil {
		return fmt.Errorf("errors.router.init_wizard_failed: %w", err)
	}
	if !completed {
		return fmt.Errorf("errors.router.init_cancelled")
	}
	return nil
}

// executeBuildPlan runs a BuildPlan with UI progress tracking.
// This is the adapter-layer responsibility: the core returns the plan, the router executes it.
func (r *Router) executeBuildPlan(ctx context.Context, plan *build.BuildPlan) error {
	if plan == nil {
		return nil
	}

	// Create progress tracker with lifecycle steps from the plan
	lifecycleSteps := make([]ports.LifecycleStep, len(plan.Steps))
	for i, step := range plan.Steps {
		lifecycleSteps[i] = ports.LifecycleStep{
			Title:  step.Title,
			Detail: step.Detail,
			Status: ports.LifecyclePending,
		}
	}

	uiAdapter := r.bld.GetUI()
	progress := ui.NewProgress(uiAdapter, plan.Title, plan.Subtitle, lifecycleSteps)
	progress.Render()

	// Execute each step with progress tracking
	for i, step := range plan.Steps {
		progress.StartStep(i, step.Title, step.Detail)
		if err := step.Exec(ctx); err != nil {
			progress.FailStep(err)
			if plan.Cleanup != nil {
				plan.Cleanup()
			}
			return err
		}
		progress.FinishStep()
	}

	// Success
	if plan.OnSuccess != nil {
		plan.OnSuccess()
	}
	progress.SetSubtitle(uiAdapter.GetText("build.success_sub", plan.Title))
	progress.Render()

	return nil
}
