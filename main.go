package main

import (
	"fmt"
	"os"

	"axiom/pkg/bunker"
	"axiom/pkg/install"
	"axiom/pkg/ui"
	"axiom/pkg/ui/styles"
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
	// Asi evitamos que una configuracion existente bloquee comandos como build.
	if command != "" && command != "init" {
		manager := bunker.NewManager(rootDir)
		if err := manager.Run(command, os.Args[2:]); err == nil {
			return
		} else if isKnownBunkerCommand(command) {
			fmt.Println(styles.GetLogo())
			fmt.Printf("❌ Error en el comando %q: %v\n", command, err)
			os.Exit(1)
		}
	}

	isInit := command == "init"
	_, err := os.Stat(".env")
	envExists := err == nil

	// Si ya existe configuracion y no estamos reconfigurando, evitamos relanzar el asistente.
	if envExists && !isInit {
		fmt.Println(styles.GetLogo())
		fmt.Printf("🛡️  AXIOM ya está configurado en: %s\n", rootDir)
		fmt.Println("Usa 'axiom init' para reconfigurar el búnker.")
		os.Exit(0)
	}

	program := tea.NewProgram(ui.NewModel(rootDir, envExists))
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error en el búnker: %v\n", err)
		os.Exit(1)
	}
}

func isKnownBunkerCommand(command string) bool {
	switch command {
	case "build":
		return true
	default:
		return false
	}
}
