package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	bunkeradapter "github.com/Alejandro-M-P/AXIOM/internal/adapters/bunker"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/runtime"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/system"
	slotui "github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/slots"
	ui "github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/views"
	"github.com/Alejandro-M-P/AXIOM/internal/core/build"
	"github.com/Alejandro-M-P/AXIOM/internal/core/bunker"
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
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

	// Detect system locale and set it for i18n (Regla 3: main.go can read env vars)
	i18n.SetLocale(os.Getenv("LANG"))

	// Determine root directory
	rootDir := resolveRootDir()

	// Create adapters (new internal structure)
	runtimeAdapter := runtime.NewPodmanAdapter()
	fsAdapter := filesystem.NewFSAdapter()
	uiAdapter := ui.NewConsoleUI()
	systemAdapter := system.NewSystemAdapter()
	gitAdapter := runtime.NewGitAdapter(fsAdapter)
	bunkerConfigurator := bunkeradapter.NewBunkerConfiguratorAdapter(fsAdapter, filepath.Join(rootDir, "configs", "assets"), uiAdapter)

	// Create managers via DI
	bunkerManager := bunker.NewManager(rootDir, runtimeAdapter, fsAdapter, uiAdapter, systemAdapter, gitAdapter, bunkerConfigurator)

	// Load slots from TOML files (in addition to init() registered slots)
	if err := slots.LoadAndRegisterSlots(rootDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Verify registry has items after loading
	if err := slots.ValidateNotEmpty(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create shell runner for command execution
	shellRunner := system.NewShellRunner()

	// Create slot manager with registry and engine
	slotRegistry := slots.GetRegistry()
	slotEngine := slots.NewInstallerEngine(slotRegistry, shellRunner)
	slotManager := slots.NewSlotManager(slotRegistry, slotEngine, uiAdapter, fsAdapter, shellRunner)

	// Create build slot adapter and wire it up
	buildSlotAdapter := newBuildSlotAdapter(slotManager, uiAdapter)

	// Create build installer (adapter layer — handles pacman, Ollama, etc.)
	buildInstaller := runtime.NewBuildInstaller(runtimeAdapter)

	// Create build manager with the slot adapter and installer
	buildManager := build.NewManager(runtimeAdapter, fsAdapter, uiAdapter, systemAdapter, "axiom-build", buildSlotAdapter, buildInstaller)

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

	// Try current working directory first (works when running from source)
	if cwd, err := os.Getwd(); err == nil {
		// Check if this looks like the project root
		if _, err := os.Stat(filepath.Join(cwd, "internal", "core", "slots")); err == nil {
			return cwd
		}
	}

	// Fall back to executable directory
	if exec, err := os.Executable(); err == nil {
		dir := filepath.Dir(exec)
		if _, err := os.Stat(filepath.Join(dir, "internal", "core", "slots")); err == nil {
			return dir
		}
	}

	return "."
}
