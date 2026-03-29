package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"axiom/internal/adapters/system"
	"axiom/internal/adapters/system/gpu"
	"axiom/internal/adapters/ui/components"
	"axiom/internal/adapters/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// BaseModel is embedded in Model for window size handling
type BaseModel struct {
	Width  int
	Height int
}

// Init sends a WindowSizeMsg to get initial dimensions
func (m *BaseModel) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{}
	}
}

// Update handles WindowSizeMsg to update dimensions
func (m *BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return nil, nil
}

type Step int

const (
	StepLanguage Step = iota
	StepConfirm
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
	BaseModel    // Embed BaseModel for Width/Height from WindowSizeMsg
	step         Step
	input        textinput.Model
	config       install.Config
	axiomPath    string
	cursor       int
	reviewCursor int
	reviewMode   bool
	detectedGPU  gpu.GPUInfo
	language     string
	envExists    bool
}

func NewModel(axiomPath string, envExists bool, lang string) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = " ❯ "

	hw := gpu.Detect()

	// Siempre empezamos preguntando el idioma
	start := StepLanguage

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
		language:    lang,
		envExists:   envExists,
	}
}

// Init initializes the model and requests window size
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.BaseModel.Init())
}

// Step devuelve el paso actual del wizard (usado para verificar estado final)
func (m Model) Step() Step {
	return m.step
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size messages from BaseModel
	m.BaseModel.Update(msg)

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
			case StepLanguage:
				if m.cursor == 0 {
					m.language = "en"
				} else {
					m.language = "es"
				}
				m.cursor = 0
				if m.envExists {
					m.step = StepConfirm
				} else {
					m.step = StepGitUser
				}

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
			case StepLanguage, StepConfirm, StepAuthMode, StepRocmMode:
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
			case StepLanguage, StepConfirm, StepAuthMode, StepRocmMode:
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
			if m.step == StepLanguage || m.step == StepConfirm || m.step == StepAuthMode || m.step == StepRocmMode {
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
			}
		}
	}

	if m.step != StepLanguage && m.step != StepConfirm && m.step != StepAuthMode && m.step != StepRocmMode && m.step != StepReview {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the form using CenteredContainer for fullscreen TUI.
func (m Model) View() string {
	header := styles.GetLogo() + "\n\n"

	if m.step == StepFinalizing {
		body := styles.GreenStyle.Render("🛡️  AXIOM: Bunker configured successfully!") + "\n\n" +
			"🚀 Next step: Run " + styles.GreenStyle.Render("axiom build") + "\n\n" +
			styles.ExampleStyle.Render("(Press Enter to exit)")

		// Use CenteredContainer for fullscreen centering
		centered := components.NewCenteredContainer(m.Width, m.Height)
		return centered.Render(styles.WindowStyle.Render(header + body))
	}

	var body string
	switch m.step {
	case StepLanguage:
		body = m.renderBox("LANGUAGE / IDIOMA", "Select your language / Selecciona tu idioma:", "English", "Español")
	case StepConfirm:
		body = m.renderBox("WARNING", "Overwrite existing .env file?", "YES, CONTINUE", "NO, EXIT")
	case StepGitUser:
		body = m.renderInput("GITHUB USER", "Enter your username:", "e.g.: user")
	case StepGitEmail:
		body = m.renderInput("GITHUB EMAIL", "Enter your email:", "e.g.: user@example.com")
	case StepAuthMode:
		body = m.renderBox("AUTHENTICATION", "Select connection type:", "SSH (Recommended)", "HTTPS (Token)")
	case StepGitToken:
		body = m.renderInput("GITHUB TOKEN", "Paste your PAT Token:", "Required for HTTPS")
	case StepBaseDir:
		body = m.renderInput("BASE DIRECTORY", "Root path:", "Enter for: "+m.config.BaseDir)
	case StepModelsDir:
		body = m.renderInput("OLLAMA MODELS", "Models location?", "Current: "+m.config.ModelsDir)
	case StepGfxVersion:
		infoGPU := fmt.Sprintf("Detected: %s", m.detectedGPU.Name)
		sugerido := "Suggested: " + m.detectedGPU.GfxVal
		if m.detectedGPU.GfxVal == "" {
			sugerido = "GFX not required"
		}
		body = m.renderInput("GPU HARDWARE", infoGPU, sugerido+" (Enter to confirm)")
	case StepRocmMode:
		sug := "Host (Recommended AMD/Intel)"
		if m.detectedGPU.Type == "nvidia" {
			sug = "Image (Recommended NVIDIA)"
		}
		body = m.renderBox("GPU DRIVERS", "Driver handling ("+sug+"):", "Host (Lightweight)", "Image (Isolated)")
	case StepReview:
		body = m.renderReview()
	}

	// Get EscapeButton text from i18n
	var escapeText string
	if Prompts != nil && Prompts["escape_button"] != nil {
		if text, ok := Prompts["escape_button"]["text"]; ok {
			escapeText = text
		}
	}
	if escapeText == "" {
		// fallback based on locale
		if currentLocale == "en" {
			escapeText = "Exit"
		} else {
			escapeText = "Salir"
		}
	}

	// Add footer with EscapeButton
	footer := "\n" + styles.ExampleStyle.Render("[Esc] "+escapeText+"  |  [Enter] Confirm")

	// Use CenteredContainer for fullscreen centering
	centered := components.NewCenteredContainer(m.Width, m.Height)
	return centered.Render(styles.WindowStyle.Render(header + body + footer))
}

func (m Model) finalizeAction() tea.Cmd {
	return func() tea.Msg {
		// Guardamos el idioma seleccionado en el entorno actual para que otras fases lo lean
		os.Setenv("AXIOM_LANG", m.language)

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
	lines = append(lines, styles.HeaderStyle.Render("FINAL SUMMARY"))
	lines = append(lines, "")
	lines = append(lines, "Review the data. Press Enter on a field to edit it.")
	lines = append(lines, styles.ExampleStyle.Render("Use Up/Down Arrows to navigate."))
	lines = append(lines, "")

	for i, item := range items {
		prefix := "  "
		if m.reviewCursor == i {
			prefix = "❯ "
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "")
	guardar := styles.InactiveButton.Render("SAVE AND CREATE")
	cancelar := styles.InactiveButton.Render("CANCEL")
	if m.reviewCursor == reviewSave {
		guardar = styles.ActiveButton.Render("SAVE AND CREATE")
	}
	if m.reviewCursor == reviewCancel {
		cancelar = styles.ActiveButton.Render("CANCEL")
	}
	lines = append(lines, guardar+"  "+cancelar)

	return strings.Join(lines, "\n")
}

func safeValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(empty)"
	}
	return v
}

func displayGFX(v string) string {
	if strings.TrimSpace(v) == "" {
		return "Not required"
	}
	return v
}

func maskToken(token, authMode string) string {
	if authMode != "https" {
		return "N/A"
	}
	if token == "" {
		return "(empty)"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
