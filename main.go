package main

import (
	"fmt"
	"os"

	"axiom/pkg/install"
	"axiom/pkg/ui"
	"axiom/pkg/ui/styles" // Lo mantenemos para el mensaje de aviso
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := install.CheckDeps(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dir, _ := os.Getwd()
	isInit := len(os.Args) > 1 && os.Args[1] == "init"
	
	_, err := os.Stat(".env")
	envExists := err == nil

	// LÓGICA DE SALIDA LIMPIA:
	if envExists && !isInit {
		// Solo aquí imprimimos el logo, porque no lanzamos la UI
		fmt.Println(styles.GetLogo()) 
		fmt.Printf("🛡️  AXIOM ya está configurado en: %s\n", dir)
		fmt.Println("Usa 'axiom init' para reconfigurar el búnker.")
		os.Exit(0)
	}

	// PARA EL INSTALADOR:
	// No imprimimos nada aquí. El logo ya vive dentro de m.View() en form.go
	p := tea.NewProgram(ui.NewModel(dir, envExists))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error en el búnker: %v\n", err)
		os.Exit(1)
	}
}