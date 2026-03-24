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
}

func (i bunkerListItem) Title() string       { return i.name }
func (i bunkerListItem) Description() string { return i.detail }
func (i bunkerListItem) FilterValue() string { return i.name }

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

func selectBunkerInteractive(title, action string, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("no hay búnkeres disponibles")
	}

	items := make([]list.Item, 0, len(names))
	for _, name := range names {
		items = append(items, bunkerListItem{name: name, detail: action})
	}

	delegate := list.NewDefaultDelegate()
	picker := list.New(items, delegate, 78, 18)
	picker.Title = title
	picker.SetShowStatusBar(false)
	picker.SetFilteringEnabled(true)
	picker.Styles.Title = styles.BunkerTitleStyle
	picker.Styles.HelpStyle = styles.BunkerSubtitleStyle
	picker.Styles.NoItems = styles.BunkerWarningStyle

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
