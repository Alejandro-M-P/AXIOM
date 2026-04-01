package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/system"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/system/gpu"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/components"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/styles"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
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
	config       system.Config
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

	initialConfig := system.Config{
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
		body := styles.GreenStyle.Render("🛡️  "+i18n.GetWizardText("finalizing", "success_title")) + "\n\n" +
			"🚀 " + i18n.GetWizardText("finalizing", "next_step") + " " + styles.GreenStyle.Render("axiom build") + "\n\n" +
			styles.ExampleStyle.Render(i18n.GetWizardText("finalizing", "press_enter"))

		// Use CenteredContainer for fullscreen centering
		centered := components.NewCenteredContainer(m.Width, m.Height)
		return centered.Render(styles.WindowStyle.Render(header + body))
	}

	var body string
	switch m.step {
	case StepLanguage:
		body = m.renderBox(
			i18n.GetWizardText("step_language", "title"),
			i18n.GetWizardText("step_language", "question"),
			i18n.GetWizardText("step_language", "option_english"),
			i18n.GetWizardText("step_language", "option_spanish"),
		)
	case StepConfirm:
		body = m.renderBox(
			i18n.GetWizardText("step_confirm", "title"),
			i18n.GetWizardText("step_confirm", "question"),
			i18n.GetWizardText("step_confirm", "option_yes"),
			i18n.GetWizardText("step_confirm", "option_no"),
		)
	case StepGitUser:
		body = m.renderInput(
			i18n.GetWizardText("step_git_user", "title"),
			i18n.GetWizardText("step_git_user", "label"),
			i18n.GetWizardText("step_git_user", "help"),
		)
	case StepGitEmail:
		body = m.renderInput(
			i18n.GetWizardText("step_git_email", "title"),
			i18n.GetWizardText("step_git_email", "label"),
			i18n.GetWizardText("step_git_email", "help"),
		)
	case StepAuthMode:
		body = m.renderBox(
			i18n.GetWizardText("step_auth_mode", "title"),
			i18n.GetWizardText("step_auth_mode", "question"),
			i18n.GetWizardText("step_auth_mode", "option_ssh"),
			i18n.GetWizardText("step_auth_mode", "option_https"),
		)
	case StepGitToken:
		body = m.renderInput(
			i18n.GetWizardText("step_git_token", "title"),
			i18n.GetWizardText("step_git_token", "label"),
			i18n.GetWizardText("step_git_token", "help"),
		)
	case StepBaseDir:
		body = m.renderInput(
			i18n.GetWizardText("step_base_dir", "title"),
			i18n.GetWizardText("step_base_dir", "label"),
			i18n.GetWizardText("step_base_dir", "help_prefix")+": "+m.config.BaseDir,
		)
	case StepModelsDir:
		body = m.renderInput(
			i18n.GetWizardText("step_models_dir", "title"),
			i18n.GetWizardText("step_models_dir", "label"),
			i18n.GetWizardText("step_models_dir", "help_prefix")+": "+m.config.ModelsDir,
		)
	case StepGfxVersion:
		infoGPU := fmt.Sprintf("%s %s", i18n.GetWizardText("step_gpu", "detected_prefix"), GetTextLocalized(m.detectedGPU.Name))
		sugerido := i18n.GetWizardText("step_gpu", "suggested_prefix") + ": " + m.detectedGPU.GfxVal
		if m.detectedGPU.GfxVal == "" {
			sugerido = i18n.GetWizardText("step_gpu", "gfx_not_required")
		}
		body = m.renderInput(
			i18n.GetWizardText("step_gpu", "title"),
			infoGPU,
			sugerido+" "+i18n.GetWizardText("step_gpu", "enter_confirm"),
		)
	case StepRocmMode:
		sug := i18n.GetWizardText("step_rocm", "host_recommended_amd")
		if m.detectedGPU.Type == "nvidia" {
			sug = i18n.GetWizardText("step_rocm", "image_recommended_nvidia")
		}
		body = m.renderBox(
			i18n.GetWizardText("step_rocm", "title"),
			i18n.GetWizardText("step_rocm", "question_prefix")+sug+i18n.GetWizardText("step_rocm", "question_suffix"),
			i18n.GetWizardText("step_rocm", "option_host"),
			i18n.GetWizardText("step_rocm", "option_image"),
		)
	case StepReview:
		body = m.renderReview()
	}

	// Add footer with EscapeButton
	footer := "\n" + styles.ExampleStyle.Render(i18n.GetWizardText("common", "footer_esc_key")+" "+i18n.GetEscapeButtonText()+"  |  "+i18n.GetWizardText("common", "footer_confirm"))

	// Use CenteredContainer for fullscreen centering
	centered := components.NewCenteredContainer(m.Width, m.Height)
	return centered.Render(styles.WindowStyle.Render(header + body + footer))
}

func (m Model) finalizeAction() tea.Cmd {
	return func() tea.Msg {
		// Guardamos el idioma seleccionado en el entorno actual para que otras fases lo lean
		os.Setenv("AXIOM_LANG", m.language)

		_ = system.NewSystemAdapter().CheckDeps()
		_ = system.PrepareFS(m.axiomPath, m.config.BaseDir)
		_ = m.config.Save(m.axiomPath)
		_ = system.CreateWrapper(m.axiomPath)
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
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_git_user"), safeValue(m.config.GitUser)),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_git_email"), safeValue(m.config.GitEmail)),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_auth_mode"), safeValue(strings.ToUpper(m.config.AuthMode))),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_git_token"), maskToken(m.config.GitToken, m.config.AuthMode)),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_base_dir"), safeValue(m.config.BaseDir)),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_models_dir"), safeValue(m.config.ModelsDir)),
		fmt.Sprintf("%s %s | %s %s | %s %s",
			i18n.GetWizardText("review", "label_gpu"), safeValue(GetTextLocalized(m.detectedGPU.Name)),
			i18n.GetWizardText("review", "label_gpu_type"), safeValue(m.config.GpuType),
			i18n.GetWizardText("review", "label_gpu_gfx"), displayGFX(m.config.GfxVersion)),
		fmt.Sprintf("%s %s", i18n.GetWizardText("review", "label_gpu_drivers"), safeValue(m.config.RocmMode)),
	}

	var lines []string
	lines = append(lines, styles.HeaderStyle.Render(i18n.GetWizardText("review", "summary_title")))
	lines = append(lines, "")
	lines = append(lines, i18n.GetWizardText("review", "instruction"))
	lines = append(lines, styles.ExampleStyle.Render(i18n.GetWizardText("review", "navigate")))
	lines = append(lines, "")

	for i, item := range items {
		prefix := "  "
		if m.reviewCursor == i {
			prefix = "❯ "
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "")
	guardar := styles.InactiveButton.Render(i18n.GetWizardText("review", "btn_save"))
	cancelar := styles.InactiveButton.Render(i18n.GetWizardText("review", "btn_cancel"))
	if m.reviewCursor == reviewSave {
		guardar = styles.ActiveButton.Render(i18n.GetWizardText("review", "btn_save"))
	}
	if m.reviewCursor == reviewCancel {
		cancelar = styles.ActiveButton.Render(i18n.GetWizardText("review", "btn_cancel"))
	}
	lines = append(lines, guardar+"  "+cancelar)

	return strings.Join(lines, "\n")
}

func safeValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return i18n.GetWizardText("common", "empty")
	}
	return v
}

func displayGFX(v string) string {
	if strings.TrimSpace(v) == "" {
		return i18n.GetWizardText("common", "not_required")
	}
	return v
}

func maskToken(token, authMode string) string {
	if authMode != "https" {
		return i18n.GetWizardText("common", "not_applicable")
	}
	if token == "" {
		return i18n.GetWizardText("common", "empty")
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
