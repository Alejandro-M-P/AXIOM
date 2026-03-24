package ui

import (
	"fmt"
	"os"

	"axiom/pkg/gpu"
	"axiom/pkg/install"
	"axiom/pkg/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Step representa cada etapa del formulario de instalación
type Step int

const (
	StepConfirm   Step = iota // Confirmación de sobreescritura
	StepGitUser               // Usuario de GitHub
	StepGitEmail              // Email de GitHub
	StepAuthMode              // Modo de conexión (SSH/HTTPS)
	StepGitToken              // Token personal (si es HTTPS)
	StepBaseDir               // Directorio raíz de trabajo
	StepModelsDir             // Ruta para modelos de IA
	StepGfxVersion            // Versión de hardware gráfico (AMD/GFX)
	StepRocmMode              // Gestión de drivers (Host/Image)
	StepFinalizing            // Procesamiento de instalación
)

// Model contiene el estado completo de la TUI
type Model struct {
	step        Step            // Paso actual
	input       textinput.Model // Componente de entrada de texto
	config      install.Config  // Configuración acumulada
	axiomPath   string          // Ruta del binario axiom
	cursor      int             // Índice para selecciones (0 o 1)
	detectedGPU gpu.GPUInfo     // Información detectada por el kernel
}

// NewModel inicializa el estado con detección de hardware en tiempo real
func NewModel(axiomPath string, envExists bool) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = " ❯ "

	// Interrogamos al sistema de archivos /sys (Host) para detectar la GPU
	hw := gpu.Detect()

	start := StepGitUser
	if envExists {
		start = StepConfirm
	}

	// Valores por defecto inteligentes basados en el hardware
	initialConfig := install.Config{
		BaseDir:    fmt.Sprintf("%s/dev", os.Getenv("HOME")),
		GfxVersion: hw.GfxVal,
		RocmMode:   "host", // Por defecto en Arch Linux (rendimiento nativo)
	}

	// Ajuste específico para NVIDIA (requiere aislamiento de drivers)
	if hw.Type == "nvidia" {
		initialConfig.RocmMode = "image"
	}

	return Model{
		step:        start,
		input:       ti,
		axiomPath:   axiomPath,
		detectedGPU: hw,
		config:      initialConfig,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update gestiona la lógica de estados y eventos de teclado
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.step == StepFinalizing {
				return m, tea.Quit
			}

			val := m.input.Value()
			switch m.step {
			case StepConfirm:
				if m.cursor == 1 {
					return m, tea.Quit
				}
				m.step = StepGitUser
				m.cursor = 0

			case StepGitUser:
				if val == "" { return m, nil }
				m.config.GitUser = val
				m.step = StepGitEmail
				m.input.SetValue("")

			case StepGitEmail:
				if val == "" { return m, nil }
				m.config.GitEmail = val
				m.step = StepAuthMode

			case StepAuthMode:
				if m.cursor == 0 {
					m.config.AuthMode = "ssh"
					m.step = StepBaseDir
				} else {
					m.config.AuthMode = "https"
					m.step = StepGitToken
				}
				m.cursor = 0

			case StepGitToken:
				m.config.GitToken = val
				m.step = StepBaseDir

			case StepBaseDir:
				if val != "" { m.config.BaseDir = val }
				m.config.ModelsDir = fmt.Sprintf("%s/ai_config/models", m.config.BaseDir)
				m.step = StepModelsDir
				m.input.SetValue("")

			case StepModelsDir:
				if val != "" { m.config.ModelsDir = val }
				m.step = StepGfxVersion
				m.input.SetValue("")

			case StepGfxVersion:
				if val != "" { m.config.GfxVersion = val }
				m.step = StepRocmMode

			case StepRocmMode:
				if m.cursor == 1 {
					m.config.RocmMode = "image"
				} else {
					m.config.RocmMode = "host"
				}
				m.step = StepFinalizing
				return m, m.finalizeAction()
			}

		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			// Lógica de alternancia para botones de selección
			if m.step == StepConfirm || m.step == StepAuthMode || m.step == StepRocmMode {
				if m.cursor == 0 { m.cursor = 1 } else { m.cursor = 0 }
			}
		}
	}

	// Actualizamos el componente de input solo si no estamos en un paso de selección
	if m.step != StepConfirm && m.step != StepAuthMode && m.step != StepRocmMode {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renderiza la interfaz gráfica en la terminal
func (m Model) View() string {
	header := styles.GetLogo() + "\n\n"

	// Pantalla de finalización
	if m.step == StepFinalizing {
		body := styles.GreenStyle.Render("🛡️  AXIOM: ¡Búnker configurado correctamente!") + "\n\n" +
			"🚀 Próximo paso: Ejecuta " + styles.GreenStyle.Render("axiom build") + " para levantar el búnker." + "\n\n" +
			styles.ExampleStyle.Render("(Presiona Enter para salir)")
		return styles.WindowStyle.Render(header + body)
	}

	var body string
	switch m.step {
	case StepConfirm:
		body = m.renderBox("ADVERTENCIA", "¿Sobrescribir archivo .env existente?", "SÍ, CONTINUAR", "NO, SALIR")
	case StepGitUser:
		body = m.renderInput("USUARIO GITHUB", "Introduce tu usuario:", "ej: User")
	case StepGitEmail:
		body = m.renderInput("EMAIL GITHUB", "Introduce tu correo:", "ej: user@example.com")
	case StepAuthMode:
		body = m.renderBox("AUTENTICACIÓN", "Selecciona conexión:", "SSH (Recomendado)", "HTTPS (Token)")
	case StepGitToken:
		body = m.renderInput("TOKEN GITHUB", "Pega tu Token PAT:", "Necesario para HTTPS")
	case StepBaseDir:
		body = m.renderInput("DIRECTORIO BASE", "Ruta raíz:", "Enter para: "+m.config.BaseDir)
	case StepModelsDir:
		body = m.renderInput("MODELOS OLLAMA", "¿Ubicación de modelos?", "Actual: "+m.config.ModelsDir)
	case StepGfxVersion:
		infoGpu := fmt.Sprintf("Detectado: %s", m.detectedGPU.Name)
		sugerido := "Sugerido: " + m.detectedGPU.GfxVal
		if m.detectedGPU.GfxVal == "" { sugerido = "No se requiere GFX" }
		body = m.renderInput("GPU HARDWARE", infoGpu, sugerido+" (Enter para confirmar)")
	
	case StepRocmMode:
		// Lógica de recomendación de driver dinámica
		sug := "Host (Recomendado AMD/Intel)"
		if m.detectedGPU.Type == "nvidia" {
			sug = "Image (Recomendado NVIDIA)"
		} else if m.detectedGPU.Type == "intel" {
			sug = "Host (Recomendado Intel)"
		}

		body = m.renderBox(
			"",                             // Título vacío para un look más limpio
			"Manejo de drivers ("+sug+"):", 
			"Host (Ligero)", 
			"Image (Aislado)",
		)
	}

	return styles.WindowStyle.Render(header + body)
}

// finalizeAction ejecuta las tareas de sistema una vez completado el formulario
func (m Model) finalizeAction() tea.Cmd {
	return func() tea.Msg {
		_ = install.CheckDeps()
		_ = install.PrepareFS(m.axiomPath, m.config.BaseDir)
		_ = m.config.Save(m.axiomPath)
		_ = install.CreateWrapper(m.axiomPath)
		return nil
	}
}

// renderInput centraliza el estilo de los campos de texto
func (m Model) renderInput(titulo, etiqueta, ayuda string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s",
		styles.HeaderStyle.Render(titulo),
		etiqueta,
		styles.ExampleStyle.Render(ayuda),
		m.input.View(),
	)
}

// renderBox centraliza el estilo de las cajas de selección binaria
func (m Model) renderBox(titulo, pregunta, o1, o2 string) string {
	b1, b2 := styles.InactiveButton.Render(o1), styles.InactiveButton.Render(o2)
	if m.cursor == 0 {
		b1 = styles.ActiveButton.Render(o1)
	} else {
		b2 = styles.ActiveButton.Render(o2)
	}

	// Evitamos saltos de línea si no hay título
	if titulo == "" {
		return fmt.Sprintf("%s\n\n%s  %s", pregunta, b1, b2)
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s  %s",
		styles.HeaderStyle.Render(titulo),
		pregunta,
		b1,
		b2,
	)
}