package ui

import (
	"strings"

	"axiom/pkg/bunker"
	"axiom/pkg/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	commandKey string
	fields     []bunker.Field
	items      []string
	question   string
	cursor     int // 0 = Yes, 1 = No
	result     bool
	canceled   bool
}

func newConfirmModel(commandKey string, fields []bunker.Field, items []string, question string) confirmModel {
	return confirmModel{
		commandKey: commandKey,
		fields:     fields,
		items:      items,
		question:   question,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit

		case tea.KeyEnter:
			m.result = (m.cursor == 0) // Yes
			return m, tea.Quit

		case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown:
			if m.cursor == 0 {
				m.cursor = 1
			} else {
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	cmdData, ok := Commands[m.commandKey]
	if !ok {
		cmdData = map[string]string{"title": m.commandKey, "subtitle": "", "footer": ""}
	}

	var details []styles.BunkerDetail
	for _, f := range m.fields {
		details = append(details, styles.BunkerDetail{Label: f.Label, Value: f.Value})
	}

	lines := styles.BuildCardLines(cmdData["title"], cmdData["subtitle"], details, m.items, cmdData["footer"])
	lines = append(lines, "")
	lines = append(lines, styles.BunkerWarningStyle.Render(m.question))
	lines = append(lines, "")

	yesButton := styles.InactiveButton.Render("SÍ, CONTINUAR")
	noButton := styles.InactiveButton.Render("NO, CANCELAR")
	if m.cursor == 0 {
		yesButton = styles.ActiveButton.Render("SÍ, CONTINUAR")
	} else {
		noButton = styles.ActiveButton.Render("NO, CANCELAR")
	}
	lines = append(lines, yesButton+"  "+noButton)

	style := styles.BunkerCardStyle
	if m.commandKey == "reset" || m.commandKey == "delete" {
		style = styles.BunkerDangerCardStyle
	}

	return styles.GetLogo() + "\n" + style.Render(strings.Join(lines, "\n"))
}

func renderFormButtons(cursor int) string {
	yesButton := styles.InactiveButton.Render("SÍ, CONTINUAR")
	noButton := styles.InactiveButton.Render("NO, CANCELAR")
	if cursor == 0 {
		yesButton = styles.ActiveButton.Render("SÍ, CONTINUAR")
	} else {
		noButton = styles.ActiveButton.Render("NO, CANCELAR")
	}
	return yesButton + "  " + noButton
}

// --- FLUJO MULTIPASO PARA DELETE ---

type deleteFormModel struct {
	fields     []bunker.Field
	step       int
	cursor     int
	textInput  textinput.Model
	confirm    bool
	reason     string
	deleteCode bool
	canceled   bool
}

func newDeleteFormModel(fields []bunker.Field) deleteFormModel {
	ti := textinput.New()
	ti.Prompt = " ❯ "
	ti.Placeholder = "Escribe aquí tu justificación..."
	return deleteFormModel{fields: fields, textInput: ti, cursor: 1} // Por seguridad, el cursor empieza en NO
}

func (m deleteFormModel) Init() tea.Cmd { return nil }

func (m deleteFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.step == 0 {
				m.confirm = (m.cursor == 0)
				if !m.confirm {
					return m, tea.Quit
				}
				m.step = 1
				m.textInput.Focus()
				return m, textinput.Blink
			} else if m.step == 1 {
				m.reason = strings.TrimSpace(m.textInput.Value())
				if m.reason != "" {
					m.step = 2
					m.cursor = 1 // Por seguridad, borrar código por defecto es NO
					m.textInput.Blur()
				}
				return m, nil
			} else if m.step == 2 {
				m.deleteCode = (m.cursor == 0)
				return m, tea.Quit
			}
		case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown:
			if m.step == 0 || m.step == 2 {
				m.cursor = 1 - m.cursor // Alterna entre 0 y 1
			}
		}
	}
	if m.step == 1 {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m deleteFormModel) View() string {
	cmdData := Commands["delete"]
	var details []styles.BunkerDetail
	for _, f := range m.fields {
		details = append(details, styles.BunkerDetail{Label: f.Label, Value: f.Value})
	}

	lines := styles.BuildCardLines(cmdData["title"], cmdData["subtitle"], details, nil, cmdData["footer"])
	lines = append(lines, "")

	if m.step == 0 {
		lines = append(lines, styles.BunkerWarningStyle.Render(getPromptText("delete.confirm")))
		lines = append(lines, "")
		lines = append(lines, renderFormButtons(m.cursor))
	} else if m.step == 1 {
		lines = append(lines, styles.BunkerWarningStyle.Render(getPromptText("delete.reason")))
		lines = append(lines, m.textInput.View())
	} else if m.step == 2 {
		lines = append(lines, styles.BunkerWarningStyle.Render(getPromptText("delete.code")))
		lines = append(lines, "")
		lines = append(lines, renderFormButtons(m.cursor))
	}

	return styles.GetLogo() + "\n" + styles.BunkerDangerCardStyle.Render(strings.Join(lines, "\n"))
}

// --- FLUJO MULTIPASO PARA RESET ---

type resetFormModel struct {
	fields    []bunker.Field
	items     []string
	step      int
	cursor    int
	textInput textinput.Model
	confirm   bool
	reason    string
	canceled  bool
}

func newResetFormModel(fields []bunker.Field, items []string) resetFormModel {
	ti := textinput.New()
	ti.Prompt = " ❯ "
	ti.Placeholder = "Escribe aquí tu justificación..."
	return resetFormModel{fields: fields, items: items, textInput: ti, cursor: 1} // Empieza en NO
}

func (m resetFormModel) Init() tea.Cmd { return nil }

func (m resetFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.step == 0 {
				m.confirm = (m.cursor == 0)
				if !m.confirm {
					return m, tea.Quit
				}
				m.step = 1
				m.textInput.Focus()
				return m, textinput.Blink
			} else if m.step == 1 {
				m.reason = strings.TrimSpace(m.textInput.Value())
				if m.reason != "" {
					return m, tea.Quit
				}
				return m, nil
			}
		case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown:
			if m.step == 0 {
				m.cursor = 1 - m.cursor
			}
		}
	}
	if m.step == 1 {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m resetFormModel) View() string {
	cmdData := Commands["reset"]
	var details []styles.BunkerDetail
	for _, f := range m.fields {
		details = append(details, styles.BunkerDetail{Label: f.Label, Value: f.Value})
	}

	lines := styles.BuildCardLines(cmdData["title"], cmdData["subtitle"], details, m.items, cmdData["footer"])
	lines = append(lines, "")

	if m.step == 0 {
		lines = append(lines, styles.BunkerWarningStyle.Render(getPromptText("reset.confirm")))
		lines = append(lines, "")
		lines = append(lines, renderFormButtons(m.cursor))
	} else if m.step == 1 {
		lines = append(lines, styles.BunkerWarningStyle.Render(getPromptText("reset.reason")))
		lines = append(lines, m.textInput.View())
	}

	return styles.GetLogo() + "\n" + styles.BunkerDangerCardStyle.Render(strings.Join(lines, "\n"))
}