package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// createFormModel es el modelo Bubbletea para el formulario de creación de bunkers.
type createFormModel struct {
	step       int // 0=select image, 1=enter name, 2=confirm
	images     []string
	imageIndex int
	name       string
	nameInput  textinput.Model
	confirm    bool
	done       bool
	canceled   bool
	width      int
	height     int
}

// newCreateFormModel crea un nuevo modelo para el formulario de creación.
func newCreateFormModel(images []string) createFormModel {
	ti := textinput.New()
	ti.Prompt = " ❯ "
	ti.Placeholder = "nombre-del-bunker"
	ti.Focus()
	return createFormModel{
		step:       0,
		images:     images,
		imageIndex: 0,
		nameInput:  ti,
		confirm:    false,
	}
}

// Init inicializa el modelo.
func (m createFormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update maneja los mensajes.
func (m createFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit

		case tea.KeyEnter:
			switch m.step {
			case 0:
				// Confirmar selección de imagen
				m.step = 1
				m.nameInput.Focus()
				return m, textinput.Blink

			case 1:
				// Confirmar nombre
				m.name = strings.TrimSpace(m.nameInput.Value())
				if m.name == "" {
					return m, nil // No avanzar si está vacío
				}
				m.step = 2
				m.nameInput.Blur()
				return m, nil

			case 2:
				// Confirmar creación
				m.done = true
				m.confirm = true
				return m, tea.Quit
			}

		case tea.KeyUp, tea.KeyLeft:
			if m.step == 0 && m.imageIndex > 0 {
				m.imageIndex--
			}

		case tea.KeyDown, tea.KeyRight:
			if m.step == 0 && m.imageIndex < len(m.images)-1 {
				m.imageIndex++
			}
		}
	}

	// Update text input en paso 1
	if m.step == 1 {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renderiza el formulario.
func (m createFormModel) View() string {
	// Usar dimensiones de la terminal si están disponibles
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	windowStyle := styles.WindowStyleWithDimensions(width, height)
	t := styles.GetTheme()

	var content strings.Builder

	switch m.step {
	case 0:
		// Selección de imagen
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render("SELECT IMAGE"))
		content.WriteString("\n\n")
		content.WriteString("Choose the base image for your bunker:\n\n")

		for i, img := range m.images {
			prefix := "  "
			style := styles.ContentStyleWithDimensions(width, height).Foreground(t.Muted)
			if i == m.imageIndex {
				prefix = "❯ "
				style = styles.ContentStyleWithDimensions(width, height).Foreground(t.Primary).Bold(true)
			}
			// Mostrar nombre legible de la imagen
			displayName := getImageDisplayName(img)
			content.WriteString(style.Render(prefix+displayName) + "\n")
		}

		content.WriteString("\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render("↑/↓: Navigate | Enter: Confirm | Esc: Cancel"))

	case 1:
		// Ingreso de nombre
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render("BUNKER NAME"))
		content.WriteString("\n\n")
		content.WriteString("Enter a name for your bunker:\n\n")

		namePreview := m.nameInput.Value()
		if namePreview == "" {
			namePreview = "nombre-del-bunker"
		}
		nameStyle := styles.ContentStyleWithDimensions(width, height).Foreground(t.Text)
		content.WriteString(nameStyle.Render("❯ "+namePreview) + "\n")

		content.WriteString("\n")
		content.WriteString(m.nameInput.View() + "\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render("Enter: Confirm | Esc: Back"))

	case 2:
		// Confirmación final
		selectedImage := m.images[m.imageIndex]
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render("CONFIRM"))
		content.WriteString("\n\n")
		content.WriteString("Create bunker with these settings?\n\n")

		infoStyle := styles.ContentStyleWithDimensions(width, height).Foreground(t.Text)
		content.WriteString(infoStyle.Render(fmt.Sprintf("  Name:   %s", m.name)) + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("  Image:  %s", getImageDisplayName(selectedImage))) + "\n")

		content.WriteString("\n\n")

		// Botones
		yesBtn := styles.InactiveButton.Render("CREATE")
		noBtn := styles.InactiveButton.Render("CANCEL")
		if m.confirm {
			yesBtn = styles.ActiveButton.Render("CREATE")
		} else {
			noBtn = styles.ActiveButton.Render("CANCEL")
		}
		content.WriteString(yesBtn + "  " + noBtn + "\n")

		content.WriteString("\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render("Enter: Confirm | Esc: Back"))
	}

	return styles.GetLogo() + "\n" + windowStyle.Render(content.String())
}

// getImageDisplayName devuelve un nombre legible para mostrar la imagen.
func getImageDisplayName(image string) string {
	switch image {
	case "axiom-dev":
		return "axiom-dev (Development environment)"
	case "axiom-data":
		return "axiom-data (Data & Databases)"
	case "axiom-sandbox":
		return "axiom-sandbox (Sandbox/Testing)"
	default:
		return image
	}
}

// AskCreateBunker ejecuta el formulario TUI de creación de bunkers.
// Retorna (name, image, confirmed, error)
func (c *ConsoleUI) AskCreateBunker(images []string) (string, string, bool, error) {
	model := newCreateFormModel(images)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	finalModel, err := p.Run()

	// Cleanup terminal - exit alternate screen mode
	fmt.Print("\033[?1049l")
	fmt.Print("\033[?25h")
	os.Stdout.Sync()

	if err != nil {
		return "", "", false, fmt.Errorf("failed to run create form: %w", err)
	}

	resultModel, ok := finalModel.(createFormModel)
	if !ok {
		return "", "", false, fmt.Errorf("unexpected model type: %T", finalModel)
	}

	if resultModel.canceled {
		return "", "", false, nil
	}

	if !resultModel.done {
		return "", "", false, nil
	}

	return resultModel.name, resultModel.images[resultModel.imageIndex], true, nil
}
