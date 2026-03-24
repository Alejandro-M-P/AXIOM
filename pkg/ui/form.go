package ui

import (
	"fmt"
	"os"
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
	step Step; input textinput.Model; config install.Config; axiomPath string; cursor int
}

func NewModel(axiomPath string, envExists bool) Model {
	ti := textinput.New(); ti.Focus(); ti.Prompt = " ❯ "
	start := StepGitUser
	if envExists { start = StepConfirm }
	return Model{step: start, input: ti, axiomPath: axiomPath, config: install.Config{BaseDir: fmt.Sprintf("%s/dev", os.Getenv("HOME"))}}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc: return m, tea.Quit
		case tea.KeyEnter:
			if m.step == StepFinalizing { return m, tea.Quit }
			val := m.input.Value()
			switch m.step {
			case StepConfirm:
				if m.cursor == 1 { return m, tea.Quit }
				m.step = StepGitUser; m.cursor = 0
			case StepGitUser:
				if val == "" { return m, nil }; m.config.GitUser = val; m.step = StepGitEmail; m.input.SetValue("")
			case StepGitEmail:
				if val == "" { return m, nil }; m.config.GitEmail = val; m.step = StepAuthMode
			case StepAuthMode:
				if m.cursor == 0 { m.config.AuthMode = "ssh"; m.step = StepBaseDir } else { m.config.AuthMode = "https"; m.step = StepGitToken }
				m.cursor = 0
			case StepGitToken: m.config.GitToken = val; m.step = StepBaseDir
			case StepBaseDir:
				if val != "" { m.config.BaseDir = val }; m.config.ModelsDir = fmt.Sprintf("%s/ai_config/models", m.config.BaseDir); m.step = StepModelsDir; m.input.SetValue("")
			case StepModelsDir:
				if val != "" { m.config.ModelsDir = val }; m.step = StepGfxVersion; m.input.SetValue("")
			case StepGfxVersion: m.config.GfxVersion = val; m.step = StepRocmMode
			case StepRocmMode:
				if m.cursor == 1 { m.config.RocmMode = "image" } else { m.config.RocmMode = "host" }
				m.step = StepFinalizing; return m, m.finalizeAction()
			}
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			if m.step == StepConfirm || m.step == StepAuthMode || m.step == StepRocmMode {
				if m.cursor == 0 { m.cursor = 1 } else { m.cursor = 0 }
			}
		}
	}
	if m.step != StepConfirm && m.step != StepAuthMode && m.step != StepRocmMode {
		var cmd tea.Cmd; m.input, cmd = m.input.Update(msg); return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	header := styles.GetLogo() + "\n\n"

	// 1. Caso especial: Finalización
	if m.step == StepFinalizing {
		body := styles.GreenStyle.Render("🛡️  AXIOM: ¡Búnker configurado correctamente!") + "\n\n" +
			"🚀 Próximo paso: Ejecuta " + styles.GreenStyle.Render("axiom build") + " para montar el primer búnker base." + "\n\n" +
			styles.ExampleStyle.Render("(Presiona Enter para salir y empezar)")
		
		return styles.WindowStyle.Render(header + body)
	}

	// 2. Declaración correcta de la variable
	var body string

	// 3. Switch con llaves correctas
	switch m.step {
	case StepConfirm:
		body = m.renderBox("ADVERTENCIA", "¿Sobrescribir archivo .env?", "SÍ, CONTINUAR", "NO, SALIR")
	case StepGitUser:
		body = m.renderInput("USUARIO GITHUB", "Introduce tu usuario:", "ej: user")
	case StepGitEmail:
		body = m.renderInput("EMAIL GITHUB", "Introduce tu correo:", "ej: user@example.com")
	case StepAuthMode:
		body = m.renderBox("AUTENTICACIÓN", "Selecciona conexión:", "SSH", "HTTPS (Token)")
	case StepGitToken:
		body = m.renderInput("TOKEN GITHUB", "Pega tu Token PAT:", "Requerido para HTTPS")
	case StepBaseDir:
		body = m.renderInput("DIRECTORIO BASE", "Ruta raíz:", "Enter para: "+m.config.BaseDir)
	case StepModelsDir:
		body = m.renderInput("MODELOS OLLAMA", "¿Ubicación de modelos?", "Actual: "+m.config.ModelsDir)
	case StepGfxVersion:
		body = m.renderInput("GPU HARDWARE", "Versión GFX (Opcional):", "ej: 11.0.0")
	case StepRocmMode:
		body = m.renderBox("MODO GPU", "Manejo de drivers:", "Host (Ligero)", "Image (Aislado)")
	}

	// 4. Renderizado final fuera del switch
	return styles.WindowStyle.Render(header + body)
}

func (m Model) renderInput(t, l, h string) string { return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", styles.HeaderStyle.Render(t), l, styles.ExampleStyle.Render(h), m.input.View()) }
func (m Model) renderBox(t, q, o1, o2 string) string {
	b1, b2 := styles.InactiveButton.Render(o1), styles.InactiveButton.Render(o2)
	if m.cursor == 0 { b1 = styles.ActiveButton.Render(o1) } else { b2 = styles.ActiveButton.Render(o2) }
	return fmt.Sprintf("%s\n\n%s\n\n%s  %s", styles.HeaderStyle.Render(t), q, b1, b2)
}

func (m Model) finalizeAction() tea.Cmd { return func() tea.Msg { _ = install.CheckDeps(); _ = install.PrepareFS(m.axiomPath, m.config.BaseDir); _ = m.config.Save(m.axiomPath); _ = install.CreateWrapper(m.axiomPath); return nil } }