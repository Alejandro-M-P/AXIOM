package main

import (
	"fmt"
	"os"

	"axiom/pkg/bunker"
	"axiom/pkg/install"
	"axiom/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := install.CheckDeps(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootDir, _ := os.Getwd()
	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Cualquier comando operativo distinto de init se delega primero al orquestador.
	// Así evitamos que una configuración existente bloquee comandos como build o create.
	if command != "" && command != "init" {
		console := ui.NewConsoleUI() // Inyectamos la Interfaz visual al Core
		manager := bunker.NewManager(rootDir)
		manager.UI = console
		if err := manager.Run(command, os.Args[2:]); err == nil {
			return
		} else if bunker.KnownCommand(command) {
			console.ShowLogo()
			fmt.Println(ui.RenderCommandError(command, err))
			os.Exit(1)
		}
	}

	isInit := command == "init"
	_, err := os.Stat(".env")
	envExists := err == nil

	// Si ya existe configuración y no estamos reconfigurando, evitamos relanzar el asistente.
	if envExists && !isInit {
		console := ui.NewConsoleUI()
		console.ShowLogo()
		console.ShowLog("cli.configured", rootDir)
		console.ShowLog("cli.use_init")
		os.Exit(0)
	}

	program := tea.NewProgram(ui.NewModel(rootDir, envExists))
	if _, err := program.Run(); err != nil {
		console := ui.NewConsoleUI()
		console.ShowLog("cli.error", err)
		os.Exit(1)
	}
}
