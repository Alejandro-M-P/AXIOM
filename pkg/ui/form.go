package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"axiom/pkg/gpu"
	"axiom/pkg/install"
	"axiom/pkg/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Step int

const (
	StepConfirm Step = iota
	StepGitUser
	StepGitEmail
	StepAuthMode
	StepGitToken
	StepBaseDir
	StepModelsDir
	StepGfxVersion
	StepRocmMode
	StepFinalizing
)

type Model struct {
	step        Step
	input       textinput.Model
	config      install.Config
	axiomPath   string
	cursor      int
	detectedGPU gpu.GPUInfo
}

func NewModel(axiomPath string, envExists bool) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = " ❯ "

	hw := gpu.Detect()

	start := StepGitUser
	if envExists { start = StepConfirm }

	initialConfig := install.Config{
		BaseDir:    fmt.Sprintf("%s/dev", os.Getenv("HOME")),
		GfxVersion: hw.GfxVal, // Cargamos el valor detectado por defecto
		GpuType:    hw.Type,   // Cargamos el tipo de GPU
		RocmMode:   "host",
	}

	if hw.Type == "nvidia" { initialConfig.RocmMode = "image" }

	return Model{
		step:        start,
		input:       ti,
		axiomPath:   axiomPath,
		detectedGPU: hw,
		config:      initialConfig,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.step == StepFinalizing { return m, tea.Quit }

			val := m.input.Value()
			switch m.step {
			case StepConfirm:
				if m.cursor == 1 { return m, tea.Quit }
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
				if val != "" {
					// 🛡️ BLINDAJE: Convertir ruta relativa a absoluta en $HOME
					if !strings.HasPrefix(val, "/") {
						home, _ := os.UserHomeDir()
						val = filepath.Join(home, val)
					}
					m.config.BaseDir = val
				}
				m.config.ModelsDir = filepath.Join(m.config.BaseDir, "ai_config/models")
				m.step = StepModelsDir
				m.input.SetValue("")

			case StepModelsDir:
				if val != "" {
					if !strings.HasPrefix(val, "/") {
						home, _ := os.UserHomeDir()
						val = filepath.Join(home, val)
					}
					m.config.ModelsDir = val
				}
				m.step = StepGfxVersion
				m.input.SetValue("")

			case StepGfxVersion:
				// Si val está vacío, se queda con el hw.GfxVal inicializado en NewModel
				if val != "" { m.config.GfxVersion = val }
				m.step = StepRocmMode
				m.input.SetValue("")

			case StepRocmMode:
				if m.cursor == 1 {
					m.config.RocmMode = "image"
				} else {
					m.config.RocmMode = "host"
				}
				// 🔗 Sincronización final del tipo de GPU
				m.config.GpuType = m.detectedGPU.Type
				m.step = StepFinalizing
				return m, m.finalizeAction()
			}

		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			if m.step == StepConfirm || m.step == StepAuthMode || m.step == StepRocmMode {
				if m.cursor == 0 { m.cursor = 1 } else { m.cursor = 0 }
			}
		}
	}

	if m.step != StepConfirm && m.step != StepAuthMode && m.step != StepRocmMode {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	header := styles.GetLogo() + "\n\n"

	if m.step == StepFinalizing {
		body := styles.GreenStyle.Render("🛡️  AXIOM: ¡Búnker configurado correctamente!") + "\n\n" +
			"🚀 Próximo paso: Ejecuta " + styles.GreenStyle.Render("axiom build") + "\n\n" +
			styles.ExampleStyle.Render("(Presiona Enter para salir)")
		return styles.WindowStyle.Render(header + body)
	}

	var body string
	switch m.step {
	case StepConfirm:
		body = m.renderBox("ADVERTENCIA", "¿Sobrescribir archivo .env existente?", "SÍ, CONTINUAR", "NO, SALIR")
	case StepGitUser:
		body = m.renderInput("USUARIO GITHUB", "Introduce tu usuario:", "ej: Alejandro-M-P")
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
		sug := "Host (Recomendado AMD/Intel)"
		if m.detectedGPU.Type == "nvidia" { sug = "Image (Recomendado NVIDIA)" }
		body = m.renderBox("", "Manejo de drivers ("+sug+"):", "Host (Ligero)", "Image (Aislado)")
	}

	return styles.WindowStyle.Render(header + body)
}

func (m Model) finalizeAction() tea.Cmd {
	return func() tea.Msg {
		_ = install.CheckDeps()
		_ = install.PrepareFS(m.axiomPath, m.config.BaseDir)
		_ = m.config.Save(m.axiomPath)
		_ = install.CreateWrapper(m.axiomPath)
		return nil
	}
}

func (m Model) renderInput(titulo, etiqueta, ayuda string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", styles.HeaderStyle.Render(titulo), etiqueta, styles.ExampleStyle.Render(ayuda), m.input.View())
}

func (m Model) renderBox(titulo, pregunta, o1, o2 string) string {
	b1, b2 := styles.InactiveButton.Render(o1), styles.InactiveButton.Render(o2)
	if m.cursor == 0 { b1 = styles.ActiveButton.Render(o1) } else { b2 = styles.ActiveButton.Render(o2) }
	head := ""
	if titulo != "" { head = styles.HeaderStyle.Render(titulo) + "\n\n" }
	return fmt.Sprintf("%s%s\n\n%s  %s", head, pregunta, b1, b2)
}