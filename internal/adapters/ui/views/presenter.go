package ui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/components"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/styles"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
	tea "github.com/charmbracelet/bubbletea"
)

// GetTextLocalized returns a localized string from i18n or returns the key if not found.
// Handles keys like "fields.name", "labels.status", "logs.gpu.forced_by_env", etc.
func GetTextLocalized(key string) string {
	parts := strings.SplitN(key, ".", 2)

	if len(parts) == 2 {
		section, subkey := parts[0], parts[1]

		// Check Commands
		if section == "commands" {
			if text, ok := i18n.Commands[subkey]; ok {
				if label, ok := text["label"]; ok {
					return label
				}
				if title, ok := text["title"]; ok {
					return title
				}
			}
		}

		// Check Lifecycle
		if section == "labels" || section == "fields" {
			if text, ok := i18n.Lifecycle[subkey]; ok {
				if label, ok := text["label"]; ok {
					return label
				}
			}
			// Try as direct key in Commands
			if text, ok := i18n.Commands[section]; ok {
				if val, ok := text[subkey]; ok {
					return val
				}
			}
		}

		// Check Logs
		if text, ok := i18n.Logs[section]; ok {
			if val, ok2 := text[subkey]; ok2 {
				return val
			}
		}

		// Check Errors
		if text, ok := i18n.Errors[section]; ok {
			if val, ok2 := text[subkey]; ok2 {
				return val
			}
		}
	}

	// Fallback: try as direct key in Commands
	if text, ok := i18n.Commands[key]; ok {
		if label, ok := text["label"]; ok {
			return label
		}
		if title, ok := text["title"]; ok {
			return title
		}
	}

	// Return key if not found
	return key
}

// ConsoleUI implementa bunker.UI para pintar en la terminal
type ConsoleUI struct{}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{}
}

func (c *ConsoleUI) ShowLogo() {
	fmt.Println(styles.GetLogo())
}

func (c *ConsoleUI) ShowCommandCard(commandKey string, fields []ports.Field, items []string) {
	cmdData, ok := i18n.Commands[commandKey]
	if !ok {
		cmdData = map[string]string{"title": commandKey, "subtitle": "", "footer": ""}
	}

	// Convert fields to CardField for fullscreen TUI
	var cardFields []components.CardField
	for _, f := range fields {
		// Get localized label
		label := GetTextLocalized(f.Label)
		value := f.Value
		cardFields = append(cardFields, components.CardField{Label: label, Value: value})
	}

	// Get localized title and subtitle
	title := cmdData["title"]
	subtitle := cmdData["subtitle"]
	footer := cmdData["footer"]

	// Run fullscreen TUI
	_ = components.RunCardTUI(title, subtitle, cardFields, items, footer)
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

	// Handle 3-level keys like "errors.ui.no_tty"
	if len(parts) == 3 {
		section, cat, sub := parts[0], parts[1], parts[2]
		if section == "errors" {
			if text, ok := i18n.Errors[cat][sub]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(text, args...)
				}
				return text
			}
		}
		// Handle slots.xxx.name and slots.xxx.description
		if section == "slots" {
			if slotData, ok := i18n.Slots[cat]; ok {
				if text, ok := slotData[sub]; ok {
					if len(args) > 0 {
						return fmt.Sprintf(text, args...)
					}
					return text
				}
			}
		}
	}

	if len(parts) == 2 {
		cat, sub := parts[0], parts[1]
		// Check Lifecycle first
		if text, ok := i18n.Lifecycle[cat][sub]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
		// Then check Commands (for slot_wizard, etc.)
		if text, ok := i18n.Commands[cat][sub]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}

	// Handle 3-level lifecycle keys like "build.commit.author" → lifecycle["build_commit"]["author"]
	if len(parts) == 3 {
		section, cat, sub := parts[0], parts[1], parts[2]
		if section == "build" {
			combined := section + "_" + cat
			if text, ok := i18n.Lifecycle[combined][sub]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(text, args...)
				}
				return text
			}
		}
		// Handle slots.xxx.name and slots.xxx.description
		if section == "slots" {
			if slotData, ok := i18n.Slots[cat]; ok {
				if text, ok := slotData[sub]; ok {
					if len(args) > 0 {
						return fmt.Sprintf(text, args...)
					}
					return text
				}
			}
		}
		// Handle errors.xxx.yyy
		if section == "errors" {
			if text, ok := i18n.Errors[cat][sub]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(text, args...)
				}
				return text
			}
		}
	}
	// Fallback to missing translation placeholder
	if text, ok := i18n.Errors["ui"]["missing_translation"]; ok {
		return text
	}
	return "[?]"
}

func (c *ConsoleUI) ClearScreen() {
	clearScreen()
}

// clearScreen centralizes terminal control codes for screen clearing.
// These are NOT user-visible text — they're ANSI escape sequences for terminal control.
func clearScreen() {
	fmt.Print("\033[2J\033[H\033[3J")
	fmt.Print("\r\x1b[2K")
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
	// Convert fields to CardField for fullscreen TUI
	var cardFields []components.CardField
	for _, f := range fields {
		label := GetTextLocalized(f.Label)
		cardFields = append(cardFields, components.CardField{Label: label, Value: f.Value})
	}

	// Get localized title and subtitle
	localizedTitle := GetTextLocalized(title)
	localizedSubtitle := GetTextLocalized(subtitle)
	localizedFooter := GetTextLocalized(footer)

	// Run fullscreen TUI
	_ = components.RunCardTUI(localizedTitle, localizedSubtitle, cardFields, items, localizedFooter)
}

func (c *ConsoleUI) ShowLog(logKey string, args ...any) {
	parts := strings.Split(logKey, ".")
	if len(parts) == 2 {
		if text, ok := i18n.Logs[parts[0]][parts[1]]; ok {
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
		cat, sub := parts[0], parts[1]
		// Check Prompts first
		if text, ok := i18n.Prompts[cat][sub]; ok {
			return text
		}
		// Fallback to Commands
		if text, ok := i18n.Commands[cat][sub]; ok {
			return text
		}
	}
	// Fallback to missing translation placeholder
	if text, ok := i18n.Errors["ui"]["missing_translation"]; ok {
		return text
	}
	return "[?]"
}

func (c *ConsoleUI) ShowHelp() {
	cmdData := i18n.Commands["help"]
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
	c.ShowLog("logs.ui.init_not_implemented")
	return nil
}

// RunInitWizardResult executes the initialization wizard and returns whether it completed successfully.
func (c *ConsoleUI) RunInitWizardResult(ctx context.Context) (bool, error) {
	c.ShowLog("logs.ui.init_not_implemented")
	return false, nil
}

// RunInitWizardWithParams executes the initialization wizard with specific parameters.
func (c *ConsoleUI) RunInitWizardWithParams(ctx context.Context, axiomPath string, envExists bool, lang string, homeDir string) (bool, error) {
	model := NewModel(axiomPath, envExists, lang, homeDir)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("errors.ui.wizard_failed: %w", err)
	}
	// Check if the wizard completed successfully
	// If the step is StepFinalizing, the user saved the configuration
	// If it's any other step, the user cancelled
	resultModel := finalModel.(Model)
	if resultModel.Step() != StepFinalizing {
		return false, errors.New("errors.ui.init_cancelled")
	}
	return true, nil
}

// RunFullscreen runs a Bubbletea model in fullscreen mode using the alternate screen.
// This is the central TUI runner that ensures consistent fullscreen behavior across all TUI components.
// Returns the final model and any error encountered during execution.
func RunFullscreen(model tea.Model) (tea.Model, error) {
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("errors.ui.tui_failed: %w", err)
	}

	return finalModel, nil
}

// RunFullscreenSimple runs a Bubbletea model in fullscreen mode without stdin/stdout customization.
// Use this for simpler cases where default input/output is sufficient.
func RunFullscreenSimple(model tea.Model) (tea.Model, error) {
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("errors.ui.tui_failed: %w", err)
	}

	return finalModel, nil
}

// GetBunkerVolumeFlags devuelve los flags de volumen para crear un contenedor.
