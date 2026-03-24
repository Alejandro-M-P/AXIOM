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
	StepReview
	StepFinalizing
)

const (
	reviewGitUser = iota
	reviewGitEmail
	reviewAuthMode
	reviewGitToken
	reviewBaseDir
	reviewModelsDir
	reviewGpu
	reviewRocmMode
	reviewSave
	reviewCancel
)

type Model struct {
	step         Step
	input        textinput.Model
	config       install.Config
	axiomPath    string
	cursor       int
	reviewCursor int
	reviewMode   bool
	detectedGPU  gpu.GPUInfo
}

func NewModel(axiomPath string, envExists bool) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = " ❯ "

	hw := gpu.Detect()

	start := StepGitUser
	if envExists {
		start = StepConfirm
	}

	initialConfig := install.Config{
		BaseDir:    fmt.Sprintf("%s/dev", os.Getenv("HOME")),
		GfxVersion: hw.GfxVal,
		GpuType:    hw.Type,
		RocmMode:   "host",
	}

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

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.step == StepReview {
				m.reviewCursor = reviewCancel
				return m, nil
			}
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
				if val == "" {
					return m, nil
				}
				m.config.GitUser = val
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepGitEmail

			case StepGitEmail:
				if val == "" {
					return m, nil
				}
				m.config.GitEmail = val
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepAuthMode

			case StepAuthMode:
				if m.cursor == 0 {
					m.config.AuthMode = "ssh"
					m.config.GitToken = ""
					m.cursor = 0
					if m.reviewMode {
						return m.backToReview(), nil
					}
					m.step = StepBaseDir
				} else {
					m.config.AuthMode = "https"
					m.cursor = 0
					m.step = StepGitToken
					m.input.SetValue(m.config.GitToken)
				}

			case StepGitToken:
				m.config.GitToken = val
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepBaseDir

			case StepBaseDir:
				oldBaseDir := m.config.BaseDir
				if val != "" {
					if !strings.HasPrefix(val, "/") {
						home, _ := os.UserHomeDir()
						val = filepath.Join(home, val)
					}
					m.config.BaseDir = val
				}
				oldDefaultModelsDir := filepath.Join(oldBaseDir, "ai_config/models")
				if m.config.ModelsDir == "" || m.config.ModelsDir == oldDefaultModelsDir {
					m.config.ModelsDir = filepath.Join(m.config.BaseDir, "ai_config/models")
				}
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepModelsDir

			case StepModelsDir:
				if val != "" {
					if !strings.HasPrefix(val, "/") {
						home, _ := os.UserHomeDir()
						val = filepath.Join(home, val)
					}
					m.config.ModelsDir = val
				}
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepGfxVersion

			case StepGfxVersion:
				if val != "" {
					m.config.GfxVersion = val
				}
				m.input.SetValue("")
				if m.reviewMode {
					return m.backToReview(), nil
				}
				m.step = StepRocmMode

			case StepRocmMode:
				if m.cursor == 1 {
					m.config.RocmMode = "image"
				} else {
					m.config.RocmMode = "host"
				}
				m.config.GpuType = m.detectedGPU.Type
				m.cursor = 0
				m.step = StepReview

			case StepReview:
				switch m.reviewCursor {
				case reviewGitUser:
					m = m.startTextEdit(StepGitUser, m.config.GitUser)
				case reviewGitEmail:
					m = m.startTextEdit(StepGitEmail, m.config.GitEmail)
				case reviewAuthMode:
					m.reviewMode = true
					m.step = StepAuthMode
					if m.config.AuthMode == "https" {
						m.cursor = 1
					} else {
						m.cursor = 0
					}
				case reviewGitToken:
					m = m.startTextEdit(StepGitToken, m.config.GitToken)
				case reviewBaseDir:
					m = m.startTextEdit(StepBaseDir, m.config.BaseDir)
				case reviewModelsDir:
					m = m.startTextEdit(StepModelsDir, m.config.ModelsDir)
				case reviewGpu:
					m = m.startTextEdit(StepGfxVersion, m.config.GfxVersion)
				case reviewRocmMode:
					m.reviewMode = true
					m.step = StepRocmMode
					if m.config.RocmMode == "image" {
						m.cursor = 1
					} else {
						m.cursor = 0
					}
				case reviewSave:
					m.step = StepFinalizing
					return m, m.finalizeAction()
				case reviewCancel:
					return m, tea.Quit
				}
			}

		case tea.KeyUp:
			switch m.step {
			case StepConfirm, StepAuthMode, StepRocmMode:
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
			case StepReview:
				if m.reviewCursor == 0 {
					m.reviewCursor = reviewCancel
				} else {
					m.reviewCursor--
				}
			}

		case tea.KeyDown:
			switch m.step {
			case StepConfirm, StepAuthMode, StepRocmMode:
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
			case StepReview:
				if m.reviewCursor == reviewCancel {
					m.reviewCursor = 0
				} else {
					m.reviewCursor++
				}
			}

		case tea.KeyLeft, tea.KeyRight:
			if m.step == StepConfirm || m.step == StepAuthMode || m.step == StepRocmMode {
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
			}
		}
	}

	if m.step != StepConfirm && m.step != StepAuthMode && m.step != StepRocmMode && m.step != StepReview {
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
		infoGPU := fmt.Sprintf("Detectado: %s", m.detectedGPU.Name)
		sugerido := "Sugerido: " + m.detectedGPU.GfxVal
		if m.detectedGPU.GfxVal == "" {
			sugerido = "No se requiere GFX"
		}
		body = m.renderInput("GPU HARDWARE", infoGPU, sugerido+" (Enter para confirmar)")
	case StepRocmMode:
		sug := "Host (Recomendado AMD/Intel)"
		if m.detectedGPU.Type == "nvidia" {
			sug = "Image (Recomendado NVIDIA)"
		}
		body = m.renderBox("DRIVERS GPU", "Manejo de drivers ("+sug+"):", "Host (Ligero)", "Image (Aislado)")
	case StepReview:
		body = m.renderReview()
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

func (m Model) startTextEdit(step Step, current string) Model {
	m.reviewMode = true
	m.step = step
	m.input.SetValue(current)
	m.input.CursorEnd()
	return m
}

func (m Model) backToReview() Model {
	m.reviewMode = false
	m.step = StepReview
	m.cursor = 0
	m.input.SetValue("")
	return m
}

func (m Model) renderInput(titulo, etiqueta, ayuda string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", styles.HeaderStyle.Render(titulo), etiqueta, styles.ExampleStyle.Render(ayuda), m.input.View())
}

func (m Model) renderBox(titulo, pregunta, o1, o2 string) string {
	b1, b2 := styles.InactiveButton.Render(o1), styles.InactiveButton.Render(o2)
	if m.cursor == 0 {
		b1 = styles.ActiveButton.Render(o1)
	} else {
		b2 = styles.ActiveButton.Render(o2)
	}
	head := ""
	if titulo != "" {
		head = styles.HeaderStyle.Render(titulo) + "\n\n"
	}
	return fmt.Sprintf("%s%s\n\n%s  %s", head, pregunta, b1, b2)
}

func (m Model) renderReview() string {
	items := []string{
		fmt.Sprintf("GitHub User: %s", safeValue(m.config.GitUser)),
		fmt.Sprintf("GitHub Email: %s", safeValue(m.config.GitEmail)),
		fmt.Sprintf("Auth Mode: %s", safeValue(strings.ToUpper(m.config.AuthMode))),
		fmt.Sprintf("Git Token: %s", maskToken(m.config.GitToken, m.config.AuthMode)),
		fmt.Sprintf("Base Dir: %s", safeValue(m.config.BaseDir)),
		fmt.Sprintf("Models Dir: %s", safeValue(m.config.ModelsDir)),
		fmt.Sprintf("GPU: %s | Tipo: %s | GFX: %s", safeValue(m.detectedGPU.Name), safeValue(m.config.GpuType), displayGFX(m.config.GfxVersion)),
		fmt.Sprintf("Drivers GPU: %s", safeValue(m.config.RocmMode)),
	}

	var lines []string
	lines = append(lines, styles.HeaderStyle.Render("RESUMEN FINAL"))
	lines = append(lines, "")
	lines = append(lines, "Revisa los datos. Pulsa Enter sobre un campo para editarlo.")
	lines = append(lines, styles.ExampleStyle.Render("Usa Flecha Arriba/Abajo para moverte."))
	lines = append(lines, "")

	for i, item := range items {
		prefix := "  "
		if m.reviewCursor == i {
			prefix = "❯ "
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "")
	guardar := styles.InactiveButton.Render("GUARDAR Y CREAR")
	cancelar := styles.InactiveButton.Render("CANCELAR")
	if m.reviewCursor == reviewSave {
		guardar = styles.ActiveButton.Render("GUARDAR Y CREAR")
	}
	if m.reviewCursor == reviewCancel {
		cancelar = styles.ActiveButton.Render("CANCELAR")
	}
	lines = append(lines, guardar+"  "+cancelar)

	return strings.Join(lines, "\n")
}

func safeValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(vacío)"
	}
	return v
}

func displayGFX(v string) string {
	if strings.TrimSpace(v) == "" {
		return "No se requiere"
	}
	return v
}

func maskToken(token, authMode string) string {
	if authMode != "https" {
		return "No aplica"
	}
	if token == "" {
		return "(vacío)"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
