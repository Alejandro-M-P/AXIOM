package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/runtime"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/system"
	slotui "github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/slots"
	ui "github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/views"
	"github.com/Alejandro-M-P/AXIOM/internal/build"
	"github.com/Alejandro-M-P/AXIOM/internal/bunker"
	"github.com/Alejandro-M-P/AXIOM/internal/slots"

	// Blank imports to trigger slot item registration via init()
	// This ensures the global registry is populated when the app starts
	_ "github.com/Alejandro-M-P/AXIOM/internal/slots/data"
	_ "github.com/Alejandro-M-P/AXIOM/internal/slots/dev/ia"
	_ "github.com/Alejandro-M-P/AXIOM/internal/slots/dev/languages"
	_ "github.com/Alejandro-M-P/AXIOM/internal/slots/dev/tools"
	_ "github.com/Alejandro-M-P/AXIOM/internal/slots/sandbox"
)

func main() {
	// Signal handling — graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println()
		os.Exit(0)
	}()

	// Determine root directory
	rootDir := resolveRootDir()

	// Create adapters (new internal structure)
	runtimeAdapter := runtime.NewPodmanAdapter()
	fsAdapter := filesystem.NewFSAdapter()
	uiAdapter := ui.NewConsoleUI()
	systemAdapter := system.NewSystemAdapter()

	// Create managers via DI
	bunkerManager := bunker.NewManager(rootDir, runtimeAdapter, fsAdapter, uiAdapter, systemAdapter)

	// Load slots from TOML files (in addition to init() registered slots)
	if err := slots.LoadAndRegisterSlots(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load slots from TOML: %v\n", err)
	}

	// Verify registry has items after loading
	if err := slots.ValidateNotEmpty(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: slot registry validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create slot manager with registry and engine
	slotRegistry := slots.GetRegistry()
	slotEngine := slots.NewInstallerEngine(slotRegistry)
	slotManager := slots.NewSlotManager(slotRegistry, slotEngine, uiAdapter, fsAdapter)

	// Create build slot adapter and wire it up
	buildSlotAdapter := newBuildSlotAdapter(slotManager, uiAdapter)

	// Create build manager with the slot adapter
	buildManager := build.NewManager(runtimeAdapter, fsAdapter, uiAdapter, systemAdapter, "axiom-build", buildSlotAdapter)

	// Create slot UI adapter for router
	slotUI := slotui.NewSlotSelectorUI(slotManager, uiAdapter)

	// Create router with all managers and dispatch
	router := NewRouter(bunkerManager, buildManager, slotManager, slotUI, rootDir, fsAdapter)
	if err := router.Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveRootDir determines the root directory of the AXIOM project.
// Priority: AXIOM_PATH > executable directory > cwd.
func resolveRootDir() string {
	if p := os.Getenv("AXIOM_PATH"); p != "" {
		return p
	}

	if exec, err := os.Executable(); err == nil {
		return filepath.Dir(exec)
	}

	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return "."
}
