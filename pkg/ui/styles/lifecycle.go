package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	LifecyclePending = "pending"
	LifecycleRunning = "running"
	LifecycleDone    = "done"
	LifecycleError   = "error"
)

type LifecycleStep struct {
	Title  string
	Detail string
	Status string
}

var (
	LifecycleShellStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Cyan).
		Padding(1, 2).
		Margin(1, 2).
		Width(84)

	LifecycleTitleStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(Dark).
		Padding(0, 1).
		Bold(true)

	LifecycleSubtitleStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Italic(true)

	LifecycleStepPendingStyle = lipgloss.NewStyle().Foreground(Gray)
	LifecycleStepRunningStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	LifecycleStepDoneStyle    = lipgloss.NewStyle().Foreground(Green).Bold(true)
	LifecycleStepErrorStyle   = lipgloss.NewStyle().Foreground(Red).Bold(true)
	LifecycleMetaStyle        = lipgloss.NewStyle().Foreground(White)
	LifecyclePathStyle        = lipgloss.NewStyle().Foreground(Cyan)
)

func RenderLifecycle(title, subtitle string, steps []LifecycleStep) string {
	completed := 0
	current := 0
	for i, step := range steps {
		if step.Status == LifecycleDone {
			completed++
		}
		if step.Status == LifecycleRunning {
			current = i + 1
		}
	}
	if current == 0 && len(steps) > 0 {
		current = completed
		if current == 0 {
			current = 1
		}
	}

	var lines []string
	lines = append(lines, LifecycleTitleStyle.Render(title))
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, LifecycleSubtitleStyle.Render(subtitle))
	}
	lines = append(lines, "")
	lines = append(lines, renderProgressBar(completed, len(steps)))
	lines = append(lines, LifecycleMetaStyle.Render(fmt.Sprintf("Progreso: %d/%d", completed, len(steps))))
	lines = append(lines, "")

	for i, step := range steps {
		label := fmt.Sprintf("%d. %s", i+1, step.Title)
		if strings.TrimSpace(step.Detail) != "" {
			label += "  " + step.Detail
		}
		lines = append(lines, renderLifecycleStep(label, step.Status))
	}

	return LifecycleShellStyle.Render(strings.Join(lines, "\n"))
}

func RenderLifecycleError(title string, steps []LifecycleStep, err error, where string) string {
	var lines []string
	lines = append(lines, RenderLifecycle(title, "Error durante el ciclo de vida", steps))
	lines = append(lines, lipgloss.NewStyle().Foreground(Red).Bold(true).Render("Error:"), err.Error())
	if strings.TrimSpace(where) != "" {
		lines = append(lines, LifecycleMetaStyle.Render("Ubicacion:"), LifecyclePathStyle.Render(where))
	}
	return strings.Join(lines, "\n")
}

func renderProgressBar(done, total int) string {
	if total <= 0 {
		total = 1
	}
	width := 28
	filled := done * width / total
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	return lipgloss.NewStyle().Foreground(Cyan).Bold(true).Render("[" + bar + "]")
}

func renderLifecycleStep(label, status string) string {
	prefix := "[ ]"
	style := LifecycleStepPendingStyle

	switch status {
	case LifecycleRunning:
		prefix = "[>]"
		style = LifecycleStepRunningStyle
	case LifecycleDone:
		prefix = "[x]"
		style = LifecycleStepDoneStyle
	case LifecycleError:
		prefix = "[!]"
		style = LifecycleStepErrorStyle
	}

	return style.Render(prefix + " " + label)
}
