package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/styles"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
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
	ti.Prompt = " " + i18n.GetWizardText("create_form", "cursor_prefix")
	ti.Placeholder = i18n.GetWizardText("create_form", "placeholder_name")
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
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "select_image")))
		content.WriteString("\n\n")
		content.WriteString(i18n.GetWizardText("create_form", "choose_image") + "\n\n")

		for i, img := range m.images {
			prefix := "  "
			style := styles.ContentStyleWithDimensions(width, height).Foreground(t.Muted)
			if i == m.imageIndex {
				prefix = i18n.GetWizardText("create_form", "cursor_prefix")
				style = styles.ContentStyleWithDimensions(width, height).Foreground(t.Primary).Bold(true)
			}
			// Mostrar nombre legible de la imagen
			displayName := getImageDisplayName(img)
			content.WriteString(style.Render(prefix+displayName) + "\n")
		}

		content.WriteString("\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "navigate")))

	case 1:
		// Ingreso de nombre
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "bunker_name")))
		content.WriteString("\n\n")
		content.WriteString(i18n.GetWizardText("create_form", "enter_name") + "\n\n")

		namePreview := m.nameInput.Value()
		if namePreview == "" {
			namePreview = i18n.GetWizardText("create_form", "placeholder_name")
		}
		nameStyle := styles.ContentStyleWithDimensions(width, height).Foreground(t.Text)
		content.WriteString(nameStyle.Render(i18n.GetWizardText("create_form", "cursor_prefix")+namePreview) + "\n")

		content.WriteString("\n")
		content.WriteString(m.nameInput.View() + "\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "enter_confirm")))

	case 2:
		// Confirmación final
		selectedImage := m.images[m.imageIndex]
		content.WriteString(styles.HeaderStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "confirm_title")))
		content.WriteString("\n\n")
		content.WriteString(i18n.GetWizardText("create_form", "confirm_question") + "\n\n")

		infoStyle := styles.ContentStyleWithDimensions(width, height).Foreground(t.Text)
		nameLabel := i18n.GetWizardText("create_form", "name_label")
		imageLabel := i18n.GetWizardText("create_form", "image_label")
		content.WriteString(infoStyle.Render(fmt.Sprintf("  %s  %s", nameLabel, m.name)) + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("  %s  %s", imageLabel, getImageDisplayName(selectedImage))) + "\n")

		content.WriteString("\n\n")

		// Botones
		createBtn := i18n.GetWizardText("create_form", "create_btn")
		cancelBtn := i18n.GetWizardText("create_form", "cancel_btn")
		yesBtn := styles.InactiveButton.Render(createBtn)
		noBtn := styles.InactiveButton.Render(cancelBtn)
		if m.confirm {
			yesBtn = styles.ActiveButton.Render(createBtn)
		} else {
			noBtn = styles.ActiveButton.Render(cancelBtn)
		}
		content.WriteString(yesBtn + "  " + noBtn + "\n")

		content.WriteString("\n")
		content.WriteString(styles.FooterStyleWithDimensions(width).Render(i18n.GetWizardText("create_form", "enter_confirm")))
	}

	return styles.GetLogo() + "\n" + windowStyle.Render(content.String())
}

// getImageDisplayName devuelve un nombre legible para mostrar la imagen.
func getImageDisplayName(image string) string {
	switch image {
	case "axiom-dev":
		return i18n.GetWizardText("create_form", "image_dev")
	case "axiom-data":
		return i18n.GetWizardText("create_form", "image_data")
	case "axiom-sandbox":
		return i18n.GetWizardText("create_form", "image_sandbox")
	default:
		return image
	}
}

// cleanupTerminal forces exit from alternate screen mode and flushes stdout
// to prevent corrupted display when running Bubble Tea programs multiple times
func cleanupTerminal() {
	// Exit alternate screen mode
	fmt.Print("\033[?1049l")
	// Reset cursor visibility (show cursor)
	fmt.Print("\033[?25h")
	// Flush stdout to ensure sequences are sent
	os.Stdout.Sync()
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

	// Cleanup terminal
	cleanupTerminal()

	if err != nil {
		return "", "", false, fmt.Errorf("errors.ui.create_form_failed: %w", err)
	}

	resultModel, ok := finalModel.(createFormModel)
	if !ok {
		return "", "", false, fmt.Errorf("errors.ui.unexpected_model")
	}

	if resultModel.canceled {
		return "", "", false, nil
	}

	if !resultModel.done {
		return "", "", false, nil
	}

	return resultModel.name, resultModel.images[resultModel.imageIndex], true, nil
}
