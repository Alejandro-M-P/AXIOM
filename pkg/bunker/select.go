package bunker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"axiom/pkg/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type bunkerListItem struct {
	name   string
	detail string
	filter string
}

func (i bunkerListItem) Title() string       { return i.name }
func (i bunkerListItem) Description() string { return i.detail }
func (i bunkerListItem) FilterValue() string { return i.filter }

type bunkerPickerModel struct {
	list     list.Model
	selected string
	canceled bool
}

func (m bunkerPickerModel) Init() tea.Cmd { return nil }

func (m bunkerPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(bunkerListItem); ok {
				m.selected = item.name
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m bunkerPickerModel) View() string {
	return styles.GetLogo() + "\n" + m.list.View()
}

func selectBunkerName(names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("no hay búnkeres disponibles para eliminar")
	}

	items := make([]list.Item, 0, len(names))
	for _, name := range names {
		items = append(items, bunkerListItem{
			name:   name,
			detail: "Pulsa Enter para elegir este búnker",
			filter: name,
		})
	}

	return runBunkerPicker(items, "Selecciona un búnker", "Escribe para filtrar, usa flechas y Enter para continuar")
}

func selectBunkerRow(rows []styles.BunkerRow) (styles.BunkerRow, error) {
	if len(rows) == 0 {
		return styles.BunkerRow{}, fmt.Errorf("no hay búnkeres disponibles")
	}

	items := make([]list.Item, 0, len(rows))
	rowMap := make(map[string]styles.BunkerRow, len(rows))
	for _, row := range rows {
		rowMap[row.Name] = row
		detail := fmt.Sprintf("%s | %s | %s | %s", fallbackPickerValue(row.Status, "-"), fallbackPickerValue(row.Image, "-"), fallbackPickerValue(row.GPU, "-"), truncatePickerValue(row.ProjectPath, 30))
		filter := strings.Join([]string{row.Name, row.Status, row.Image, row.GPU, row.ProjectPath, row.EnvPath, row.GitBranch}, " ")
		items = append(items, bunkerListItem{name: row.Name, detail: detail, filter: filter})
	}

	selected, err := runBunkerPicker(items, "Búnkeres", "Busca por nombre, GPU, imagen o ruta. Enter abre la ficha")
	if err != nil {
		return styles.BunkerRow{}, err
	}
	row, ok := rowMap[selected]
	if !ok {
		return styles.BunkerRow{}, fmt.Errorf("no se pudo resolver el búnker seleccionado")
	}
	return row, nil
}

func runBunkerPicker(items []list.Item, title, help string) (string, error) {
	delegate := list.NewDefaultDelegate()
	picker := list.New(items, delegate, 92, 20)
	picker.Title = title
	picker.SetShowStatusBar(false)
	picker.SetFilteringEnabled(true)
	picker.SetShowPagination(true)
	picker.Styles.Title = styles.BunkerTitleStyle
	picker.Styles.HelpStyle = styles.BunkerSubtitleStyle
	picker.Styles.NoItems = styles.BunkerWarningStyle
	picker.Help.Styles.ShortKey = styles.BunkerValueStyle
	picker.Help.Styles.ShortDesc = styles.BunkerSubtitleStyle
	picker.Help.Styles.FullKey = styles.BunkerValueStyle
	picker.Help.Styles.FullDesc = styles.BunkerSubtitleStyle
	if strings.TrimSpace(help) != "" {
		picker.NewStatusMessage(help)
	}

	result, err := tea.NewProgram(bunkerPickerModel{list: picker}).Run()
	if err != nil {
		return "", err
	}
	finalModel, ok := result.(bunkerPickerModel)
	if !ok {
		return "", fmt.Errorf("no se pudo resolver la selección del búnker")
	}
	if finalModel.canceled || strings.TrimSpace(finalModel.selected) == "" {
		return "", fmt.Errorf("selección cancelada")
	}
	return finalModel.selected, nil
}

func listBunkerNames(cfg EnvConfig) ([]string, error) {
	seen := map[string]struct{}{}
	var names []string

	output, err := runCommandOutputQuiet("distrobox-list", "--no-color")
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			if !strings.Contains(line, "|") {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) < 2 {
				continue
			}
			name := strings.TrimSpace(parts[1])
			if name == "" || name == defaultBuildContainerName {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	entries, readErr := os.ReadDir(filepath.Join(cfg.BaseDir, ".entorno"))
	if readErr == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.TrimSpace(entry.Name())
			if name == "" || name == defaultBuildContainerName {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

func truncatePickerValue(value string, width int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= width {
		return trimmed
	}
	if width <= 1 {
		return trimmed[:width]
	}
	return trimmed[:width-1] + "…"
}

func fallbackPickerValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
