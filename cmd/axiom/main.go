package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"axiom/internal/adapters/filesystem"
	"axiom/internal/adapters/runtime"
	ui "axiom/internal/adapters/ui/views"
	"axiom/internal/build"
	"axiom/internal/bunker"
	"axiom/internal/slots"

	// Blank imports to trigger slot item registration via init()
	// This ensures the global registry is populated when the app starts
	_ "axiom/internal/slots/data"
	_ "axiom/internal/slots/dev/ia"
	_ "axiom/internal/slots/dev/languages"
	_ "axiom/internal/slots/dev/tools"
	_ "axiom/internal/slots/sandbox"
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

	// Create managers via DI
	bunkerManager := bunker.NewManager(rootDir, runtimeAdapter, fsAdapter, uiAdapter)

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
	buildManager := build.NewManager(runtimeAdapter, fsAdapter, uiAdapter, "axiom-build", buildSlotAdapter)

	// Create router with all managers and dispatch
	router := NewRouter(bunkerManager, buildManager, slotManager, rootDir, fsAdapter)
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
