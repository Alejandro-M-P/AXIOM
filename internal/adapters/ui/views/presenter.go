package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"axiom/internal/adapters/ui/styles"
	"axiom/internal/ports"
	tea "github.com/charmbracelet/bubbletea"
)

// missingTextPlaceholder is shown when a translation key is not found
const missingTextPlaceholder = "[Texto no disponible]"

// ConsoleUI implementa bunker.UI para pintar en la terminal
type ConsoleUI struct{}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{}
}

func (c *ConsoleUI) ShowLogo() {
	fmt.Println(styles.GetLogo())
}

func (c *ConsoleUI) ShowCommandCard(commandKey string, fields []ports.Field, items []string) {
	cmdData, ok := Commands[commandKey]
	if !ok {
		cmdData = map[string]string{"title": commandKey, "subtitle": "", "footer": ""}
	}

	var details []styles.BunkerDetail
	for _, f := range fields {
		details = append(details, styles.BunkerDetail{Label: f.Label, Value: f.Value})
	}

	fmt.Println(styles.RenderBunkerCard(
		cmdData["title"],
		cmdData["subtitle"],
		details,
		items,
		cmdData["footer"],
	))
}

func (c *ConsoleUI) AskConfirmInCard(commandKey string, fields []ports.Field, items []string, promptKey string) (bool, error) {
	question := getPromptText(promptKey)
	model := newConfirmModel(commandKey, fields, items, question)

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	resultModel := finalModel.(confirmModel)
	if resultModel.canceled {
		return false, nil
	}
	return resultModel.result, nil
}

func (c *ConsoleUI) AskDelete(name string, fields []ports.Field) (bool, string, bool, error) {
	model := newDeleteFormModel(fields)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, "", false, err
	}
	res := finalModel.(deleteFormModel)
	if res.canceled || !res.confirm {
		return false, "", false, nil
	}
	return res.confirm, res.reason, res.deleteCode, nil
}

func (c *ConsoleUI) AskReset(fields []ports.Field, items []string) (bool, string, error) {
	model := newResetFormModel(fields, items)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, "", err
	}
	res := finalModel.(resetFormModel)
	if res.canceled || !res.confirm {
		return false, "", nil
	}
	return res.confirm, res.reason, nil
}

func (c *ConsoleUI) GetText(key string, args ...any) string {
	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		cat, sub := parts[0], parts[1]
		// Check Lifecycle first
		if text, ok := Lifecycle[cat][sub]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
		// Then check Commands (for slot_wizard, etc.)
		if text, ok := Commands[cat][sub]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}
	// Log missing key for debugging (using fmt.Fprintln to stderr)
	fmt.Fprintf(os.Stderr, "[i18n] Missing translation key: %s\n", key)
	return missingTextPlaceholder
}

func (c *ConsoleUI) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (c *ConsoleUI) RenderLifecycle(title, subtitle string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep) {
	fmt.Println(styles.RenderLifecycleWithTasks(title, subtitle, mapSteps(steps), taskTitle, mapSteps(taskSteps)))
}

func (c *ConsoleUI) RenderLifecycleError(title string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep, err error, where string) {
	fmt.Println(styles.RenderLifecycleError(title, mapSteps(steps), taskTitle, mapSteps(taskSteps), err, where))
}

func mapSteps(steps []ports.LifecycleStep) []styles.LifecycleStep {
	var res []styles.LifecycleStep
	for _, s := range steps {
		res = append(res, styles.LifecycleStep{Title: s.Title, Detail: s.Detail, Status: s.Status})
	}
	return res
}

func (c *ConsoleUI) ShowWarning(title, subtitle string, fields []ports.Field, items []string, footer string) {
	var details []styles.BunkerDetail
	for _, f := range fields {
		details = append(details, styles.BunkerDetail{Label: f.Label, Value: f.Value})
	}
	fmt.Println(styles.RenderBunkerWarning(title, subtitle, details, items, footer))
}

func (c *ConsoleUI) ShowLog(logKey string, args ...any) {
	parts := strings.Split(logKey, ".")
	if len(parts) == 2 {
		if text, ok := Logs[parts[0]][parts[1]]; ok {
			if len(args) > 0 {
				fmt.Printf(text+"\n", args...)
			} else {
				fmt.Println(text)
			}
		}
	}
}

func (c *ConsoleUI) AskString(promptKey string) (string, error) {
	fmt.Print(getPromptText(promptKey))
	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res), nil
}

func (c *ConsoleUI) AskConfirm(promptKey string, args ...any) (bool, error) {
	prompt := getPromptText(promptKey)
	if len(args) > 0 {
		prompt = fmt.Sprintf(prompt, args...)
	}
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	res = strings.ToLower(strings.TrimSpace(res))
	return res == "s" || res == "y" || res == "si" || res == "sí" || res == "yes", nil
}

func getPromptText(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		if text, ok := Prompts[parts[0]][parts[1]]; ok {
			return text
		}
	}
	// Log missing key for debugging (using fmt.Fprintln to stderr)
	fmt.Fprintf(os.Stderr, "[i18n] Missing prompt key: %s\n", key)
	return missingTextPlaceholder
}

func (c *ConsoleUI) ShowHelp() {
	cmdData := Commands["help"]
	var details []styles.BunkerDetail
	for i := 1; i <= 10; i++ {
		if val, ok := cmdData[fmt.Sprintf("cmd_%d", i)]; ok {
			parts := strings.SplitN(val, ":", 2)
			if len(parts) == 2 {
				details = append(details, styles.BunkerDetail{Label: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
			}
		}
	}
	var items []string
	for i := 1; i <= 2; i++ {
		if val, ok := cmdData[fmt.Sprintf("item_%d", i)]; ok {
			items = append(items, val)
		}
	}
	fmt.Println(styles.RenderBunkerCard(
		cmdData["title"],
		cmdData["subtitle"],
		details,
		items,
		cmdData["footer"],
	))
}

// RunInitWizard executes the initialization wizard
func (c *ConsoleUI) RunInitWizard(ctx context.Context) error {
	// Placeholder - actual implementation would run the TUI wizard
	fmt.Println("Init wizard not yet implemented")
	return nil
}
